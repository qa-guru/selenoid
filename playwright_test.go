package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aerokube/selenoid/session"
	"github.com/stretchr/testify/assert"
)

func TestParsePlaywrightRequest(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=smoke&enableVideo=true")
	assert.NoError(t, err)

	browser, version, caps, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "playwright-chromium", browser)
	assert.Equal(t, "1.61.1", version)
	assert.Equal(t, "smoke", caps.TestName)
	assert.True(t, caps.Video)
}

func TestParsePlaywrightRequestLabels(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium/1.61.1?name=Manual+session&labels.manual=true")
	assert.NoError(t, err)

	_, _, caps, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "Manual session", caps.TestName)
	assert.Equal(t, map[string]string{"manual": "true"}, caps.Labels)
}

func TestPlaywrightSessionDeletedViaHub(t *testing.T) {
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
}

func TestParsePlaywrightRequestDefaultVersion(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/playwright-chromium")
	assert.NoError(t, err)

	browser, version, _, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "playwright-chromium", browser)
	assert.Equal(t, "", version)
}
