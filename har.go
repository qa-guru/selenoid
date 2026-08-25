package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/browser"
	"github.com/mafredri/cdp/rpcc"
	harpkg "github.com/qa-guru/selenoid/har"
	"github.com/qa-guru/selenoid/info"
	"github.com/qa-guru/selenoid/session"
)

const harFileExtension = ".har"

// har serves the HAR (HTTP Archive) artifact directory, mirroring the /video/
// and /logs/ endpoints:
//
//	GET    /har/                 HTML directory listing
//	GET    /har/?json[&...]      paginated JSON listing (limit/offset/q)
//	GET    /har/<file>.har       download a single HAR file
//	DELETE /har/<file>.har       remove a HAR file
//
// When -har-output-dir is not configured the feature is disabled and every
// request returns 404 (there is no directory to serve).
func har(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	if harOutputDir == "" {
		http.Error(w, "HAR recording is not enabled (set -har-output-dir)", http.StatusNotFound)
		return
	}
	if r.Method == http.MethodDelete {
		deleteFileIfExists(requestId, w, r, harOutputDir, paths.Har, "DELETED_HAR_FILE")
		return
	}
	user, remote := info.RequestInfo(r)
	if _, ok := r.URL.Query()[jsonParam]; ok {
		listVideosAsJson(requestId, w, r, harOutputDir)
		return
	}
	log.Printf("[%d] [HAR_LISTING] [%s] [%s]", requestId, user, remote)
	fileServer := http.StripPrefix(paths.Har, http.FileServer(http.Dir(harOutputDir)))
	fileServer.ServeHTTP(w, r)
}

// harCaptureEnabled reports whether a native hub HAR should be recorded for a
// session with the given capabilities. Recording is a Docker-only, CDP-based
// feature: it needs the browser container DevTools endpoint (port 7070), so a
// non-empty devtools host:port is required.
func harCaptureEnabled(caps session.Caps, devtoolsHostPort string) bool {
	return harOutputDir != "" && caps.HAR && !disableDocker && devtoolsHostPort != ""
}

// startHarCapture opens a CDP HAR recorder against the browser container
// DevTools page endpoint (ws://<host:7070>/page — the same target exposed by
// the se:cdp capability and the /devtools/<id>/page proxy). captureBodies opts
// into Network.getResponseBody (harContent=bodies). Returns nil on failure so
// callers can treat capture as best-effort.
func startHarCapture(requestId uint64, sessionId, devtoolsHostPort string, captureBodies bool) *harpkg.Session {
	wsURL := "ws://" + devtoolsHostPort + "/page"
	rec, err := startHarSession(wsURL, captureBodies)
	if err != nil {
		log.Printf("[%d] [HAR_CAPTURE_FAILED] [%s] [%v]", requestId, sessionId, err)
		return nil
	}
	mode := "meta"
	if captureBodies {
		mode = "bodies"
	}
	log.Printf("[%d] [HAR_CAPTURE_STARTED] [%s] [%s] [%s]", requestId, sessionId, wsURL, mode)
	return rec
}

// startHarCapturePlaywright retries page-level CDP attach. Playwright
// launchServer has no page until the client calls newPage(); the hub races a
// short retry loop so Network.enable runs on that page. Clients should create
// the page before navigating (or pause briefly after newPage) so the first
// navigation is captured — same one-writer CDP path as WebDriver /page.
func startHarCapturePlaywright(requestId uint64, sessionId, devtoolsHostPort string, captureBodies bool, attempts int, delay time.Duration) *harpkg.Session {
	var lastErr error
	for i := 0; i < attempts; i++ {
		for _, wsURL := range devtoolsPageWSURLs(devtoolsHostPort) {
			rec, err := startHarSession(wsURL, captureBodies)
			if err == nil {
				mode := "meta"
				if captureBodies {
					mode = "bodies"
				}
				log.Printf("[%d] [HAR_CAPTURE_STARTED] [%s] [%s] [%s]", requestId, sessionId, wsURL, mode)
				return rec
			}
			lastErr = err
		}
		time.Sleep(delay)
	}
	log.Printf("[%d] [HAR_CAPTURE_FAILED] [%s] [%v]", requestId, sessionId, lastErr)
	return nil
}

// ensureDevtoolsPage opens about:blank on the browser DevTools HTTP API so manual
// Playwright sessions (UI bare WebSocket, no client newPage) still expose a /page
// target for hub-side HAR capture.
func ensureDevtoolsPage(requestId uint64, sessionId, devtoolsHostPort string) {
	devtoolsHostPort = strings.TrimSpace(devtoolsHostPort)
	if devtoolsHostPort == "" {
		return
	}
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodPut, "http://"+devtoolsHostPort+"/json/new?about:blank", nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%d] [HAR_PAGE_BOOTSTRAP_FAILED] [%s] [%v]", requestId, sessionId, err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		log.Printf("[%d] [HAR_PAGE_BOOTSTRAP_FAILED] [%s] [HTTP %d]", requestId, sessionId, resp.StatusCode)
		return
	}
	log.Printf("[%d] [HAR_PAGE_BOOTSTRAP] [%s]", requestId, sessionId)
}

// fitPlaywrightWindow resizes the Chromium window on Xvfb to screenResolution.
// Manual UI sessions (and HAR /json/new) otherwise keep Chrome's default ~800×600
// because --window-size is ignored without a window manager.
func fitPlaywrightWindow(requestId uint64, sessionId, devtoolsHostPort, screenResolution string) {
	devtoolsHostPort = strings.TrimSpace(devtoolsHostPort)
	if devtoolsHostPort == "" {
		return
	}
	width, height := screenSizePixels(screenResolution)
	var lastErr error
	for i := 0; i < 40; i++ {
		if err := setDevtoolsWindowBounds(devtoolsHostPort, width, height); err == nil {
			log.Printf("[%d] [PLAYWRIGHT_WINDOW_FIT] [%s] [%dx%d]", requestId, sessionId, width, height)
			return
		} else {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}
	log.Printf("[%d] [PLAYWRIGHT_WINDOW_FIT_FAILED] [%s] [%v]", requestId, sessionId, lastErr)
}

func setDevtoolsWindowBounds(devtoolsHostPort string, width, height int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	wsURL := "ws://" + devtoolsHostPort + "/page"
	conn, err := rpcc.DialContext(ctx, wsURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := cdp.NewClient(conn)
	win, err := c.Browser.GetWindowForTarget(ctx, browser.NewGetWindowForTargetArgs())
	if err != nil {
		return err
	}
	maxErr := c.Browser.SetWindowBounds(ctx, browser.NewSetWindowBoundsArgs(win.WindowID, browser.Bounds{
		WindowState: browser.WindowStateMaximized,
	}))
	if maxErr == nil {
		return nil
	}
	left, top := 0, 0
	w, h := width, height
	return c.Browser.SetWindowBounds(ctx, browser.NewSetWindowBoundsArgs(win.WindowID, browser.Bounds{
		Left:        &left,
		Top:         &top,
		Width:       &w,
		Height:      &h,
		WindowState: browser.WindowStateNormal,
	}))
}

type devtoolsTarget struct {
	Type                 string `json:"type"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func devtoolsPageWSURLs(devtoolsHostPort string) []string {
	devtoolsHostPort = strings.TrimSpace(devtoolsHostPort)
	if devtoolsHostPort == "" {
		return nil
	}
	defaultURL := "ws://" + devtoolsHostPort + "/page"
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + devtoolsHostPort + "/json/list")
	if err != nil {
		return []string{defaultURL}
	}
	defer resp.Body.Close()
	var targets []devtoolsTarget
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return []string{defaultURL}
	}
	out := make([]string, 0, len(targets)+1)
	seen := map[string]struct{}{}
	add := func(url string) {
		url = strings.TrimSpace(url)
		if url == "" {
			return
		}
		if _, ok := seen[url]; ok {
			return
		}
		seen[url] = struct{}{}
		out = append(out, url)
	}
	for _, target := range targets {
		if target.Type == "page" {
			add(target.WebSocketDebuggerURL)
		}
	}
	add(defaultURL)
	return out
}

func startHarSession(wsURL string, captureBodies bool) (*harpkg.Session, error) {
	if captureBodies {
		return harpkg.StartWithBodies(context.Background(), wsURL)
	}
	return harpkg.Start(context.Background(), wsURL)
}

// devtoolsWsHostPort normalizes a HostPort.Devtools value (host:port) for use
// as a websocket authority. It is a thin indirection kept for readability.
func devtoolsWsHostPort(hostPort string) string {
	return strings.TrimSpace(hostPort)
}
