package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlaywrightRequest(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/chromium/1.61.1?name=smoke&enableVideo=true")
	assert.NoError(t, err)

	browser, version, caps, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "chromium", browser)
	assert.Equal(t, "1.61.1", version)
	assert.Equal(t, "smoke", caps.TestName)
	assert.True(t, caps.Video)
}

func TestParsePlaywrightRequestLabels(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/chromium/1.61.1?name=Manual+session&labels.manual=true")
	assert.NoError(t, err)

	_, _, caps, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "Manual session", caps.TestName)
	assert.Equal(t, map[string]string{"manual": "true"}, caps.Labels)
}

func TestParsePlaywrightRequestDefaultVersion(t *testing.T) {
	u, err := url.Parse("ws://localhost:4444/playwright/chromium")
	assert.NoError(t, err)

	browser, version, _, err := parsePlaywrightRequest(u)
	assert.NoError(t, err)
	assert.Equal(t, "chromium", browser)
	assert.Equal(t, "", version)
}
