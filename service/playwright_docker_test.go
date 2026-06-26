package service

import (
	"testing"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/session"
	"github.com/stretchr/testify/assert"
)

func TestPlaywrightContainerEnv(t *testing.T) {
	env := playwrightContainerEnv("3000", session.Caps{Headless: true})
	assert.Contains(t, env, "PW_PORT=3000")
	assert.Contains(t, env, "PW_HEADLESS=true")
	assert.Contains(t, env, "MANUAL_SESSION=false")

	manual := playwrightContainerEnv("3000", session.Caps{
		VNC:      true,
		Headless: false,
		TestName: "Manual session",
	})
	assert.Contains(t, manual, "MANUAL_SESSION=true")
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
	return &config.Browser{Port: port, Image: "qaguru/playwright-chromium:1.61.1"}
}
