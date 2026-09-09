package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/qa-guru/selenoid/session"
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

	t.Run("harContent=bodies with enableHAR", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableHAR=true&harContent=bodies")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.True(t, caps.HAR)
		assert.Equal(t, "bodies", caps.HARContent)
		assert.True(t, caps.HARBodies())
	})

	t.Run("harContent omit defaults to meta", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableHAR=true")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.True(t, caps.HAR)
		assert.Empty(t, caps.HARContent)
		assert.False(t, caps.HARBodies())
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

func TestPlaywrightHarRegistryWaitsForLateComplete(t *testing.T) {
	id := "har-late-complete"
	beginPlaywrightHar(id, "late.har")
	go func() {
		time.Sleep(40 * time.Millisecond)
		completePlaywrightHar(id, nil)
	}()

	h := waitPlaywrightHarDone(id, time.Second)
	assert.NotNil(t, h)
	assert.True(t, harSlotDone(h))
	assert.Equal(t, "late.har", h.name)
	assert.Equal(t, h, takePlaywrightHar(id))
	assert.Nil(t, takePlaywrightHar(id))
}

func TestPlaywrightDeleteSessionDoesNotBlockOnPendingHar(t *testing.T) {
	sessionId := "pw-pending-har"
	beginPlaywrightHar(sessionId, "pending.har")
	t.Cleanup(func() { takePlaywrightHar(sessionId) })

	wsURL, err := url.Parse("ws://127.0.0.1:3000")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.True(t, queue.Wait(ctx))
	queue.Create()
	t.Cleanup(func() {
		if _, ok := sessions.Get(sessionId); ok {
			sessions.Remove(sessionId)
			queue.Release()
		}
	})

	canceled := make(chan struct{})
	sessions.Put(sessionId, &session.Session{
		Caps: session.Caps{Name: "playwright-chromium", HAR: true},
		URL:  wsURL,
		HostPort: session.HostPort{
			Playwright: "127.0.0.1:3000",
		},
		Cancel: func() {
			close(canceled)
			completePlaywrightHar(sessionId, nil)
		},
		TimeoutCh: make(chan struct{}),
	})

	start := time.Now()
	playwrightDeleteSession(1, sessionId, "", "")
	assert.Less(t, time.Since(start), 2*time.Second)
	select {
	case <-canceled:
	default:
		t.Fatal("expected container cancel while HAR attach was still pending")
	}
	_, ok := sessions.Get(sessionId)
	assert.False(t, ok)
}

func TestPlaywrightDeleteSessionWritesHarAndRenamesLog(t *testing.T) {
	harDir := t.TempDir()
	logDir := t.TempDir()
	prevH, prevL := harOutputDir, logOutputDir
	harOutputDir, logOutputDir = harDir, logDir
	t.Cleanup(func() {
		harOutputDir, logOutputDir = prevH, prevL
		takePlaywrightHar("pw-art-session")
	})

	sessionId := "pw-art-session"
	tempLog := "selenoid-pw-temp.log"
	assert.NoError(t, os.WriteFile(filepath.Join(logDir, tempLog), []byte("pw-container-log\n"), 0644))
	putPlaywrightHar(sessionId, nil, "")

	wsURL, err := url.Parse("ws://127.0.0.1:3000")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.True(t, queue.Wait(ctx))
	queue.Create()
	t.Cleanup(func() {
		if _, ok := sessions.Get(sessionId); ok {
			sessions.Remove(sessionId)
			queue.Release()
		}
	})

	canceled := false
	sessions.Put(sessionId, &session.Session{
		Caps: session.Caps{
			Name:    "playwright-chromium",
			HAR:     true,
			Log:     true,
			LogName: tempLog,
		},
		URL: wsURL,
		HostPort: session.HostPort{
			Playwright: "127.0.0.1:3000",
		},
		Cancel:    func() { canceled = true },
		TimeoutCh: make(chan struct{}),
	})

	playwrightDeleteSession(1, sessionId, "", "")
	assert.True(t, canceled)
	_, ok := sessions.Get(sessionId)
	assert.False(t, ok)

	harBytes, err := os.ReadFile(filepath.Join(harDir, sessionId+".har"))
	assert.NoError(t, err)
	assert.Contains(t, string(harBytes), `"version": "1.2"`)

	logBytes, err := os.ReadFile(filepath.Join(logDir, sessionId+".log"))
	assert.NoError(t, err)
	assert.Equal(t, "pw-container-log\n", string(logBytes))
	assert.FileExists(t, filepath.Join(logDir, sessionId+".json"))
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

func TestParsePlaywrightRequestSocksProxy(t *testing.T) {
	t.Run("host:port becomes PW_PROXY socks5 URL", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?socksProxy=proxy.qaguru.school%3A7777")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.Contains(t, caps.Env, "PW_PROXY=socks5://proxy.qaguru.school:7777")
	})

	t.Run("full proxy URL is preserved", func(t *testing.T) {
		u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?socksProxy=http%3A%2F%2Fproxy.example%3A3128")
		assert.NoError(t, err)

		_, _, caps, err := parsePlaywrightRequest(u)
		assert.NoError(t, err)
		assert.Contains(t, caps.Env, "PW_PROXY=http://proxy.example:3128")
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
