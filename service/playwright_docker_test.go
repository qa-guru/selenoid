package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/session"
	"github.com/stretchr/testify/assert"
)

func TestPlaywrightContainerEnv(t *testing.T) {
	t.Run("Playwright container env", func(t *testing.T) {
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
	})
}

func TestGetPlaywrightPortConfigWithVNC(t *testing.T) {
	t.Run("Get playwright port config with vnc", func(t *testing.T) {
		browser := configBrowser("3000")
		cfg, err := getPlaywrightPortConfig(browser, session.Caps{VNC: true}, Environment{})
		assert.NoError(t, err)
		assert.NotEmpty(t, cfg.VNCPort)
		_, ok := cfg.ExposedPorts[cfg.VNCPort]
		assert.True(t, ok)
	})
}

func configBrowser(port string) *config.Browser {
	return &config.Browser{Port: port, Image: "qaguru/playwright-chromium:1.61.1", Protocol: playwrightProtocol}
}

func TestGetPlaywrightPortConfigWithoutVNC(t *testing.T) {
	t.Run("Get playwright port config without vnc", func(t *testing.T) {
		browser := configBrowser("3000")
		cfg, err := getPlaywrightPortConfig(browser, session.Caps{}, Environment{})
		assert.NoError(t, err)
		assert.Empty(t, cfg.VNCPort)
	})
}

func TestIsPlaywrightBrowser(t *testing.T) {
	t.Run("Is playwright browser", func(t *testing.T) {
		assert.True(t, isPlaywrightBrowser(&config.Browser{Protocol: playwrightProtocol}))
		assert.False(t, isPlaywrightBrowser(&config.Browser{Protocol: "webdriver"}))
		assert.False(t, isPlaywrightBrowser(&config.Browser{}))
	})
}

func TestWaitPlaywrightServer(t *testing.T) {
	t.Run("Wait playwright server", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		hostPort := strings.TrimPrefix(srv.URL, "http://")
		assert.NoError(t, waitPlaywrightServer(hostPort, 2*time.Second))
	})

	t.Run("Wait playwright server timeout", func(t *testing.T) {
		err := waitPlaywrightServer("127.0.0.1:1", 200*time.Millisecond)
		assert.Error(t, err)
	})
}

func TestWaitTCP(t *testing.T) {
	t.Run("Wait tcp", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		hostPort := strings.TrimPrefix(srv.URL, "http://")
		assert.NoError(t, waitTCP(hostPort, 2*time.Second))
	})
}
