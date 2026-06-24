package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aerokube/selenoid/info"
	"github.com/aerokube/selenoid/session"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var playwrightUpgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

func playwrightConnect(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	user, remote := info.RequestInfo(r)

	if !websocket.IsWebSocketUpgrade(r) {
		http.Error(w, "WebSocket upgrade required for Playwright connections", http.StatusBadRequest)
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

	sess := &session.Session{
		Quota:     user,
		Caps:      caps,
		URL:       startedService.Url,
		Container: startedService.Container,
		HostPort:  startedService.HostPort,
		Cancel:    startedService.Cancel,
		Timeout:   sessionTimeout,
		TimeoutCh: onTimeout(sessionTimeout, func() {
			playwrightDeleteSession(requestId, sessionId)
		}),
		Started: time.Now(),
	}
	sessions.Put(sessionId, sess)
	queue.Create()
	sessionCreated = true
	log.Printf("[%d] [PLAYWRIGHT_SESSION_CREATED] [%s] [%s] [%s]", requestId, sessionId, browser, version)

	clientConn, err := playwrightUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_UPGRADE_FAILED] [%s] [%v]", requestId, sessionId, err)
		playwrightDeleteSession(requestId, sessionId)
		return
	}
	defer clientConn.Close()

	backendURL := startedService.Url.String()
	log.Printf("[%d] [PLAYWRIGHT_CONNECTING] [%s] [%s]", requestId, sessionId, backendURL)
	backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		log.Printf("[%d] [PLAYWRIGHT_BACKEND_DIAL_FAILED] [%s] [%v]", requestId, sessionId, err)
		playwrightDeleteSession(requestId, sessionId)
		return
	}
	defer backendConn.Close()

	proxyPlaywrightWebSocket(requestId, sessionId, clientConn, backendConn)
	playwrightDeleteSession(requestId, sessionId)
}

func playwrightDeleteSession(requestId uint64, sessionId string) {
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
	log.Printf("[%d] [PLAYWRIGHT_SESSION_DELETED] [%s]", requestId, sessionId)
}

func proxyPlaywrightWebSocket(requestId uint64, sessionId string, clientConn, backendConn *websocket.Conn) {
	errCh := make(chan error, 2)

	go func() {
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := backendConn.WriteMessage(messageType, message); err != nil {
				errCh <- err
				return
			}
		}
	}()

	go func() {
		for {
			messageType, message, err := backendConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := clientConn.WriteMessage(messageType, message); err != nil {
				errCh <- err
				return
			}
		}
	}()

	err := <-errCh
	if err != nil && !isExpectedWSClose(err) {
		log.Printf("[%d] [PLAYWRIGHT_PROXY_ERROR] [%s] [%v]", requestId, sessionId, err)
		return
	}
	log.Printf("[%d] [PLAYWRIGHT_SESSION_CLOSED] [%s]", requestId, sessionId)
}

func isExpectedWSClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
		errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "use of closed network connection")
}

func parsePlaywrightRequest(u *url.URL) (browser, version string, caps session.Caps, err error) {
	trimmed := strings.TrimPrefix(u.Path, paths.Playwright)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", "", caps, errors.New("browser name is required in Playwright URL")
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
	if _, ok := values["enableVideo"]; ok {
		caps.Video = queryBool(values, "enableVideo")
	}
	if _, ok := values["enableLog"]; ok {
		caps.Log = queryBool(values, "enableLog")
	}
	if tz := values.Get("timeZone"); tz != "" {
		caps.TimeZone = tz
	}
	for key, vals := range values {
		if strings.HasPrefix(key, "env.") && len(vals) > 0 {
			caps.Env = append(caps.Env, fmt.Sprintf("%s=%s", strings.TrimPrefix(key, "env."), vals[0]))
		}
	}
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
