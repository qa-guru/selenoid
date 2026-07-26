package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	harpkg "github.com/aerokube/selenoid/har"
	"github.com/aerokube/selenoid/info"
	"github.com/aerokube/selenoid/session"
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
// the se:cdp capability and the /devtools/<id>/page proxy). It returns nil on
// failure so callers can treat capture as best-effort.
func startHarCapture(requestId uint64, sessionId, devtoolsHostPort string) *harpkg.Session {
	wsURL := "ws://" + devtoolsHostPort + "/page"
	rec, err := harpkg.Start(context.Background(), wsURL)
	if err != nil {
		log.Printf("[%d] [HAR_CAPTURE_FAILED] [%s] [%v]", requestId, sessionId, err)
		return nil
	}
	log.Printf("[%d] [HAR_CAPTURE_STARTED] [%s] [%s]", requestId, sessionId, wsURL)
	return rec
}

// devtoolsWsHostPort normalizes a HostPort.Devtools value (host:port) for use
// as a websocket authority. It is a thin indirection kept for readability.
func devtoolsWsHostPort(hostPort string) string {
	return strings.TrimSpace(hostPort)
}
