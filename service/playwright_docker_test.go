package service

import (
	"strings"
	"testing"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/session"
	"github.com/stretchr/testify/assert"
)

func TestBuildPlaywrightStartupScriptHeadless(t *testing.T) {
	script := buildPlaywrightStartupScript("1.52.0", "3000", "chromium", session.Caps{})
	assert.Equal(t, "exec node /home/pwuser/node_modules/playwright/cli.js run-server --port 3000 --host 0.0.0.0", script)
}

func TestBuildPlaywrightStartupScriptWithVNC(t *testing.T) {
	script := buildPlaywrightStartupScript("1.52.0", "3000", "chromium", session.Caps{
		VNC:              true,
		ScreenResolution: "1920x1080x24",
	})
	assert.True(t, strings.Contains(script, "Xvfb :99"))
	assert.True(t, strings.Contains(script, "-listen tcp"))
	assert.True(t, strings.Contains(script, "x11vnc"))
	assert.True(t, strings.Contains(script, "DISPLAY=:99"))
	assert.True(t, strings.Contains(script, "run-server"))
	assert.False(t, strings.Contains(script, "launch-headed-browser.js"))
}

func TestBuildPlaywrightStartupScriptManualHeadedVNC(t *testing.T) {
	script := buildPlaywrightStartupScript("1.52.0", "3000", "chromium", session.Caps{
		VNC:              true,
		Headless:         false,
		TestName:         "Manual session",
		ScreenResolution: "1920x1080x24",
	})
	assert.True(t, strings.Contains(script, "launch-headed-browser.js"))
	assert.True(t, strings.Contains(script, "PW_BROWSER=chromium"))
}

func TestBuildPlaywrightStartupScriptWithVideoOnly(t *testing.T) {
	script := buildPlaywrightStartupScript("1.52.0", "3000", "chromium", session.Caps{
		Video:            true,
		ScreenResolution: "1280x720x24",
	})
	assert.True(t, strings.Contains(script, "Xvfb :99"))
	assert.False(t, strings.Contains(script, "x11vnc"))
}

func TestGetPlaywrightPortConfigWithVNC(t *testing.T) {
	browser := configBrowser("3000")
	cfg, err := getPlaywrightPortConfig(browser, session.Caps{VNC: true}, Environment{})
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.VNCPort)
	_, ok := cfg.ExposedPorts[cfg.VNCPort]
	assert.True(t, ok)
}

func configBrowser(port string) *config.Browser {
	return &config.Browser{Port: port, Image: "selenoid/playwright:v1.52.0-noble"}
}