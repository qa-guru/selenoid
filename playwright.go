package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qa-guru/selenoid/event"
	harpkg "github.com/qa-guru/selenoid/har"
	"github.com/qa-guru/selenoid/info"
	"github.com/qa-guru/selenoid/session"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func playwrightConnect(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	user, remote := info.RequestInfo(r)

	if !websocket.IsWebSocketUpgrade(r) {
		http.Error(w, "WebSocket upgrade required for Playwright connections", http.StatusBadRequest)
		return
	}

	if !accessKeyOK(r) {
		log.Printf("[%d] [PLAYWRIGHT_UNAUTHORIZED] [%s] [%s]", requestId, user, remote)
		http.Error(w, "Playwright accessKey required", http.StatusUnauthorized)
		return
	}

	browser, version, caps, err := parsePlaywrightRequest(r.URL)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_BAD_REQUEST] [%s] [%s] [%v]", requestId, user, remote, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[%d] [PLAYWRIGHT_REQUEST] [%s] [%s] [%s] [%s]", requestId, user, remote, browser, version)

	if !queue.Wait(r.Context()) {
		log.Printf("[%d] [PLAYWRIGHT_CLIENT_DISCONNECTED] [%s] [%s]", requestId, user, remote)
		return
	}

	sessionCreated := false
	defer func() {
		if !sessionCreated {
			queue.Drop()
		}
	}()

	resolution, err := getScreenResolution(caps.ScreenResolution)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_BAD_SCREEN_RESOLUTION] [%s]", requestId, caps.ScreenResolution)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	caps.ScreenResolution = resolution

	videoScreenSize, err := getVideoScreenSize(caps.VideoScreenSize, resolution)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_BAD_VIDEO_SCREEN_SIZE] [%s]", requestId, caps.VideoScreenSize)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	caps.VideoScreenSize = videoScreenSize

	finalVideoName := caps.VideoName
	if caps.Video && !disableDocker {
		caps.VideoName = getTemporaryFileName(videoOutputDir, videoFileExtension)
	}

	sessionTimeout, err := getSessionTimeout(caps.SessionTimeout, maxTimeout, timeout)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_BAD_SESSION_TIMEOUT] [%s]", requestId, caps.SessionTimeout)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	starter, ok := manager.FindPlaywright(browser, version, caps, requestId)
	if !ok {
		log.Printf("[%d] [PLAYWRIGHT_ENVIRONMENT_NOT_AVAILABLE] [%s] [%s]", requestId, browser, version)
		http.Error(w, "Requested Playwright environment is not available", http.StatusBadRequest)
		return
	}

	startedService, err := starter.StartWithCancel()
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_SERVICE_STARTUP_FAILED] [%v]", requestId, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionId := uuid.NewString()
	if ggrHost != nil {
		sessionId = ggrHost.Sum() + sessionId
	}

	// Manual Playwright sessions may opt into a hub-side HAR (enableHAR query
	// param). Capture reuses the same CDP path as WebDriver — best-effort, only
	// when a DevTools endpoint is exposed for the container. Automated Playwright
	// tests should prefer the client-side recordHar option (one writer/session).
	// The recorder is stashed in a registry so whichever delete path fires
	// (client WS close, idle timeout or an explicit hub DELETE) writes the HAR.
	//
	// Playwright launchServer has no page until the client calls newPage(), so
	// HAR start runs asynchronously with retries (unlike WebDriver, where a page
	// exists before the hub returns the session).
	devtoolsHP := devtoolsWsHostPort(startedService.HostPort.Devtools)
	startPWHar := harCaptureEnabled(caps, devtoolsHP)

	sess := &session.Session{
		Quota:     user,
		Caps:      caps,
		URL:       startedService.Url,
		Container: startedService.Container,
		HostPort:  startedService.HostPort,
		Cancel:    startedService.Cancel,
		Timeout:   sessionTimeout,
		TimeoutCh: onTimeout(sessionTimeout, func() {
			playwrightDeleteSession(requestId, sessionId, finalVideoName)
		}),
		Started: time.Now(),
	}
	sessions.Put(sessionId, sess)
	queue.Create()
	sessionCreated = true
	log.Printf("[%d] [PLAYWRIGHT_SESSION_CREATED] [%s] [%s] [%s]", requestId, sessionId, browser, version)

	if startPWHar {
		harName := caps.HARName
		captureBodies := caps.HARBodies()
		go func() {
			// Manual UI sessions keep a bare WS without Playwright newPage(); seed a page
			// over DevTools HTTP so hub HAR can attach before the client navigates.
			ensureDevtoolsPage(requestId, sessionId, devtoolsHP)
			if rec := startHarCapturePlaywright(requestId, sessionId, devtoolsHP, captureBodies, 120, 250*time.Millisecond); rec != nil {
				putPlaywrightHar(sessionId, rec, harName)
			}
		}()
	}

	backendURL := startedService.Url
	log.Printf("[%d] [PLAYWRIGHT_CONNECTING] [%s] [%s]", requestId, sessionId, backendURL.String())
	proxyPlaywright(w, r, backendURL)
	playwrightDeleteSession(requestId, sessionId, finalVideoName)
}

func proxyPlaywright(w http.ResponseWriter, r *http.Request, backend *url.URL) {
	target := &url.URL{
		Scheme: "http",
		Host:   backend.Host,
		Path:   backend.Path,
	}
	if backend.Scheme == "wss" {
		target.Scheme = "https"
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = backend.Path
		req.URL.RawPath = ""
		req.URL.RawQuery = ""
	}
	proxy.ServeHTTP(w, r)
}

func isPlaywrightSession(sess *session.Session) bool {
	return sess.HostPort.Playwright != "" || (sess.URL != nil && sess.URL.Scheme == "ws")
}

func playwrightDeleteSession(requestId uint64, sessionId string, finalVideoName string) {
	sess, ok := sessions.Get(sessionId)
	if !ok {
		return
	}
	sess.Lock.Lock()
	defer sess.Lock.Unlock()
	select {
	case <-sess.TimeoutCh:
	default:
		close(sess.TimeoutCh)
	}
	sessions.Remove(sessionId)
	queue.Release()
	if sess.Cancel != nil {
		sess.Cancel()
	}
	if sess.Caps.Video && !disableDocker {
		oldVideoName := filepath.Join(videoOutputDir, sess.Caps.VideoName)
		if finalVideoName == "" {
			finalVideoName = sessionId + videoFileExtension
		}
		newVideoName := filepath.Join(videoOutputDir, finalVideoName)
		if err := os.Rename(oldVideoName, newVideoName); err != nil {
			log.Printf("[%d] [VIDEO_ERROR] [%s]", requestId, fmt.Sprintf("Failed to rename %s to %s: %v", oldVideoName, newVideoName, err))
		} else {
			event.FileCreated(event.CreatedFile{
				Event: event.Event{
					RequestId: requestId,
					SessionId: sessionId,
					Session:   sess,
				},
				Name: newVideoName,
				Type: "video",
			})
		}
	}
	if h := takePlaywrightHar(sessionId); h != nil {
		rec := h.recorder.Stop()
		finalHarName := h.name
		if finalHarName == "" {
			finalHarName = sessionId + harFileExtension
		}
		harPath := filepath.Join(harOutputDir, finalHarName)
		if err := rec.WriteFile(harPath, sess.Caps.TestName); err != nil {
			log.Printf("[%d] [HAR_ERROR] [%s]", requestId, fmt.Sprintf("Failed to write HAR %s: %v", harPath, err))
		} else {
			event.FileCreated(event.CreatedFile{
				Event: event.Event{
					RequestId: requestId,
					SessionId: sessionId,
					Session:   sess,
				},
				Name: harPath,
				Type: "har",
			})
			log.Printf("[%d] [HAR_SAVED] [%s] [%s] [%d entries]", requestId, sessionId, finalHarName, rec.EntryCount())
		}
	}
	event.SessionStopped(event.StoppedSession{
		Event: event.Event{
			RequestId: requestId,
			SessionId: sessionId,
			Session:   sess,
		},
	})
	log.Printf("[%d] [PLAYWRIGHT_SESSION_DELETED] [%s]", requestId, sessionId)
}

// playwrightHar holds a live hub HAR recorder for a manual Playwright session
// plus the requested output name. It is kept in a package-level registry so any
// session-teardown path (client WS close, idle timeout or an explicit hub
// DELETE via the /wd/hub proxy) can stop the recorder and write the HAR exactly
// once — the recorder cannot live on the closure alone because the hub-DELETE
// path removes the session before the connect handler's deferred cleanup runs.
type playwrightHar struct {
	recorder *harpkg.Session
	name     string
}

var (
	playwrightHarMu   sync.Mutex
	playwrightHarByID = map[string]*playwrightHar{}
)

// putPlaywrightHar registers a recorder for a session id.
func putPlaywrightHar(sessionId string, recorder *harpkg.Session, name string) {
	playwrightHarMu.Lock()
	defer playwrightHarMu.Unlock()
	playwrightHarByID[sessionId] = &playwrightHar{recorder: recorder, name: name}
}

// takePlaywrightHar removes and returns the recorder for a session id, or nil if
// none was registered. It is safe to call from every teardown path; only the
// first caller for a given id gets the recorder.
func takePlaywrightHar(sessionId string) *playwrightHar {
	playwrightHarMu.Lock()
	defer playwrightHarMu.Unlock()
	h := playwrightHarByID[sessionId]
	delete(playwrightHarByID, sessionId)
	return h
}

func accessKeyOK(r *http.Request) bool {
	if strings.TrimSpace(accessKeys) == "" {
		return true
	}
	provided := r.URL.Query().Get("accessKey")
	if provided == "" {
		provided = r.URL.Query().Get("access_key")
	}
	for _, key := range strings.Split(accessKeys, ",") {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) == 1 {
			return true
		}
	}
	return false
}

func parsePlaywrightRequest(u *url.URL) (browser, version string, caps session.Caps, err error) {
	trimmed := strings.TrimPrefix(u.Path, paths.Playwright)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", "", caps, fmt.Errorf("browser name is required in Playwright URL")
	}

	parts := strings.Split(trimmed, "/")
	browser = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}

	caps = session.Caps{
		Name:             browser,
		Version:          version,
		ScreenResolution: "1920x1080x24",
	}
	capsFromQuery(u.Query(), &caps)
	caps.Name = browser
	if version != "" {
		caps.Version = version
	}
	return browser, version, caps, nil
}

func capsFromQuery(values url.Values, caps *session.Caps) {
	if name := values.Get("name"); name != "" {
		caps.TestName = name
	}
	if resolution := values.Get("screenResolution"); resolution != "" {
		caps.ScreenResolution = resolution
	}
	if timeout := values.Get("sessionTimeout"); timeout != "" {
		caps.SessionTimeout = timeout
	}
	if _, ok := values["enableVNC"]; ok {
		caps.VNC = queryBool(values, "enableVNC")
	}
	if _, ok := values["headless"]; ok {
		caps.Headless = queryBool(values, "headless")
	} else {
		caps.Headless = true
	}
	if _, ok := values["enableVideo"]; ok {
		caps.Video = queryBool(values, "enableVideo")
	}
	if videoName := values.Get("videoName"); videoName != "" {
		caps.VideoName = videoName
	}
	if _, ok := values["enableLog"]; ok {
		caps.Log = queryBool(values, "enableLog")
	}
	if logName := values.Get("logName"); logName != "" {
		caps.LogName = logName
	}
	if _, ok := values["enableHAR"]; ok {
		caps.HAR = queryBool(values, "enableHAR")
	}
	if harName := values.Get("harName"); harName != "" {
		caps.HARName = harName
	}
	if harContent := values.Get("harContent"); harContent != "" {
		caps.HARContent = harContent
	}
	caps.NormalizeHARContent()
	if tz := values.Get("timeZone"); tz != "" {
		caps.TimeZone = tz
	}
	// SOCKS/HTTP proxy for launchServer / headed VNC (image reads PW_PROXY).
	// Accept host:port (→ socks5://) or a full URL with scheme.
	if proxy := values.Get("socksProxy"); proxy != "" {
		if u := normalizePlaywrightProxy(proxy); u != "" {
			caps.Env = append(caps.Env, "PW_PROXY="+u)
		}
	}
	for key, vals := range values {
		if strings.HasPrefix(key, "env.") && len(vals) > 0 {
			caps.Env = append(caps.Env, fmt.Sprintf("%s=%s", strings.TrimPrefix(key, "env."), vals[0]))
		}
		if strings.HasPrefix(key, "labels.") && len(vals) > 0 {
			if caps.Labels == nil {
				caps.Labels = make(map[string]string)
			}
			caps.Labels[strings.TrimPrefix(key, "labels.")] = vals[0]
		}
	}
}

func normalizePlaywrightProxy(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if strings.Contains(s, "://") {
		return s
	}
	return "socks5://" + s
}

func queryBool(values url.Values, key string) bool {
	raw := values.Get(key)
	if raw == "" {
		return true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return true
	}
	return parsed
}
