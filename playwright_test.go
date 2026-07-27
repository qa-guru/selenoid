package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aerokube/selenoid/session"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestParsePlaywrightRequest(t *testing.T) {
	t.Run("Parse playwright request", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=smoke&enableVideo=true")
		assert.NoError(t, err)

		browser, version, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.Equal(t, "playwright-chromium", browser)
		assert.Equal(t, "1.61.1", version)
		assert.Equal(t, "smoke", caps.TestName)
		assert.True(t, caps.Video)
	})
}

func TestParsePlaywrightRequestHar(t *testing.T) {
	t.Run("Parse playwright request enableHAR and harName", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableHAR=true&harName=manual.har")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.True(t, caps.HAR)
		assert.Equal(t, "manual.har", caps.HARName)
	})

	t.Run("enableHAR absent leaves HAR off", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=smoke")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.False(t, caps.HAR)
		assert.Empty(t, caps.HARName)
	})
}

func TestPlaywrightHarRegistryTakeOnce(t *testing.T) {
	id := "har-registry-session"
	putPlaywrightHar(id, nil, "custom.har")

	h := takePlaywrightHar(id)
	assert.NotNil(t, h)
	assert.Equal(t, "custom.har", h.name)

	// A second take for the same id yields nil so only one teardown path writes.
	assert.Nil(t, takePlaywrightHar(id))
}

func TestParsePlaywrightRequestLabels(t *testing.T) {
	t.Run("Parse playwright request labels", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=Manual+session&labels.manual=true")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.Equal(t, "Manual session", caps.TestName)
		assert.Equal(t, map[string]string{"manual": "true"}, caps.Labels)
	})
}

func TestParsePlaywrightRequestLogName(t *testing.T) {
	t.Run("Parse playwright request enableLog and logName", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableLog=true&logName=manual.log&screenResolution=1280x1024x24&timeZone=Europe%2FMoscow&env.FOO=bar")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.True(t, caps.Log)
		assert.Equal(t, "manual.log", caps.LogName)
		assert.Equal(t, "1280x1024x24", caps.ScreenResolution)
		assert.Equal(t, "Europe/Moscow", caps.TimeZone)
		assert.Equal(t, []string{"FOO=bar"}, caps.Env)
	})
}

func TestPlaywrightSessionDeletedViaHub(t *testing.T) {
	t.Run("Playwright session deleted via hub", func(t *testing.T) {
		canceled := false
		sessionId := "test-playwright-session"
		wsURL, err := url.Parse("ws://127.0.0.1:3000")
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		assert.True(t, queue.Wait(ctx))
		queue.Create()
		defer func() {
			if _, ok := sessions.Get(sessionId); ok {
				sessions.Remove(sessionId)
				queue.Release()
			}
		}()

		sessions.Put(sessionId, &session.Session{
			Caps: session.Caps{
				Name:   "playwright-chromium",
				Labels: map[string]string{"manual": "true"},
			},
			URL: wsURL,
			HostPort: session.HostPort{
				Playwright: "127.0.0.1:3000",
			},
			Cancel:    func() { canceled = true },
			TimeoutCh: make(chan struct{}),
		})

		req, err := http.NewRequest(http.MethodDelete, With(srv.URL).Path(fmt.Sprintf("/wd/hub/session/%s", sessionId)), nil)
		assert.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_, ok := sessions.Get(sessionId)
		assert.False(t, ok)
		assert.True(t, canceled)
	})
}

func TestParsePlaywrightRequestDefaultVersion(t *testing.T) {
	t.Run("Parse playwright request default version", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium")
		assert.NoError(t, err)

		browser, version, _, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.Equal(t, "playwright-chromium", browser)
		assert.Equal(t, "", version)
	})
}

func TestPlaywrightConnectRejectsNonWebSocket(t *testing.T) {
	t.Run("Playwright connect rejects non websocket", func(t *testing.T) {
		rsp, err := http.Get(With(srv.URL).Path("/playwright/playwright-chromium/1.61.1"))
		assert.NoError(t, err)
		defer rsp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, rsp.StatusCode)
	})
}

func TestPlaywrightConnectBrowserNotFound(t *testing.T) {
	t.Run("Playwright connect browser not found", func(t *testing.T) {
		manager = &BrowserNotFound{}
		wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/playwright/playwright-chromium/1.61.1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if conn != nil {
			_ = conn.Close()
		}
		assert.Error(t, err)
	})
}

func TestPlaywrightConnectProxiesWebSocket(t *testing.T) {
	t.Run("Playwright connect proxies websocket", func(t *testing.T) {
		manager = &HTTPTest{Handler: Selenium()}
		wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/playwright/playwright-chromium/1.61.1?name=smoke"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer conn.Close()

		msg := []byte(`{"id":1}`)
		err = conn.WriteMessage(websocket.TextMessage, msg)
		assert.NoError(t, err)

		_, reply, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.JSONEq(t, string(msg), string(reply))
	})
}

const playwrightPublicAccessKeyDemo = "qa_engineer:aAb_-4gs53FD"

func TestAccessKeyRequired(t *testing.T) {
	t.Run("Rejects websocket without accessKey", func(t *testing.T) {
		prev := accessKeys
		accessKeys = "user1:1234," + playwrightPublicAccessKeyDemo
		defer func() { accessKeys = prev }()

		manager = &HTTPTest{Handler: Selenium()}
		wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/playwright/playwright-chromium/1.61.1?name=smoke"
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if conn != nil {
			_ = conn.Close()
		}
		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("Accepts websocket with student accessKey", func(t *testing.T) {
		prev := accessKeys
		accessKeys = "user1:1234," + playwrightPublicAccessKeyDemo
		defer func() { accessKeys = prev }()

		manager = &HTTPTest{Handler: Selenium()}
		wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) +
			"/playwright/playwright-chromium/1.61.1?accessKey=" + url.QueryEscape("user1:1234") + "&name=smoke"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer conn.Close()
	})

	t.Run("Accepts websocket with public accessKey alias", func(t *testing.T) {
		prev := accessKeys
		accessKeys = "user1:1234, " + playwrightPublicAccessKeyDemo
		defer func() { accessKeys = prev }()

		manager = &HTTPTest{Handler: Selenium()}
		wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) +
			"/playwright/playwright-chromium/1.61.1?access_key=" + url.QueryEscape(playwrightPublicAccessKeyDemo) + "&name=smoke"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer conn.Close()
	})
}

func TestProxyPlaywright(t *testing.T) {
	t.Run("Proxy playwright to backend", func(t *testing.T) {
		upgrader := websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer c.Close()
			mt, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			_ = c.WriteMessage(mt, msg)
		}))
		defer backend.Close()

		backendHost := strings.TrimPrefix(backend.URL, "http://")
		backendURL, err := url.Parse("ws://" + backendHost + "/")
		assert.NoError(t, err)

		proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxyPlaywright(w, r, backendURL)
		}))
		defer proxySrv.Close()

		wsURL := strings.Replace(proxySrv.URL, "http://", "ws://", 1) + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer conn.Close()

		msg := []byte("ping")
		assert.NoError(t, conn.WriteMessage(websocket.TextMessage, msg))
		_, reply, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, msg, reply)
	})
}
