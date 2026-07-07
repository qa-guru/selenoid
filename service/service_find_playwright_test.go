package service

import (
	"testing"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPlaywrightManagerConfig() *config.Config {
	conf := config.NewConfig()
	conf.Browsers["chromium"] = config.Versions{
		Default: "1.61.1",
		Versions: map[string]*config.Browser{
			"1.61.1": configBrowser("3000"),
		},
	}
	conf.Browsers["firefox"] = config.Versions{
		Default: "120.0",
		Versions: map[string]*config.Browser{
			"120.0": {
				Image:    "selenoid/firefox:120.0",
				Port:     "4444",
				Protocol: "webdriver",
			},
		},
	}
	return conf
}

func newPlaywrightManager(t *testing.T, mock *playwrightDockerMock) *DefaultManager {
	t.Helper()
	return &DefaultManager{
		Environment: testPlaywrightEnvironmentPtr(t),
		Client:      mock.cli,
		Config:      testPlaywrightManagerConfig(),
	}
}

func testPlaywrightEnvironmentPtr(t *testing.T) *Environment {
	t.Helper()
	env := testPlaywrightEnvironment()
	return &env
}

func TestDefaultManagerFindPlaywright(t *testing.T) {
	caps := session.Caps{Headless: true}

	t.Run("ok", func(t *testing.T) {
		mock := newPlaywrightDockerMockWithPort(t, "127.0.0.1:3000")
		defer mock.server.Close()

		starter, ok := newPlaywrightManager(t, mock).FindPlaywright("chromium", "1.61.1", caps, 42)
		require.True(t, ok)
		require.NotNil(t, starter)

		pw, isPlaywright := starter.(*PlaywrightDocker)
		require.True(t, isPlaywright)
		assert.Equal(t, uint64(42), pw.RequestId)
		assert.Equal(t, mock.cli, pw.Client)
		assert.Equal(t, "3000", pw.Service.Port)
		assert.Equal(t, playwrightProtocol, pw.Service.Protocol)
	})

	t.Run("no client", func(t *testing.T) {
		manager := &DefaultManager{
			Environment: testPlaywrightEnvironmentPtr(t),
			Config:      testPlaywrightManagerConfig(),
		}

		starter, ok := manager.FindPlaywright("chromium", "1.61.1", caps, 42)
		assert.False(t, ok)
		assert.Nil(t, starter)
	})

	t.Run("unknown browser", func(t *testing.T) {
		mock := newPlaywrightDockerMockWithPort(t, "127.0.0.1:3000")
		defer mock.server.Close()

		starter, ok := newPlaywrightManager(t, mock).FindPlaywright("unknown", "1.0", caps, 42)
		assert.False(t, ok)
		assert.Nil(t, starter)
	})

	t.Run("not playwright protocol", func(t *testing.T) {
		mock := newPlaywrightDockerMockWithPort(t, "127.0.0.1:3000")
		defer mock.server.Close()

		starter, ok := newPlaywrightManager(t, mock).FindPlaywright("firefox", "120.0", caps, 42)
		assert.False(t, ok)
		assert.Nil(t, starter)
	})
}
