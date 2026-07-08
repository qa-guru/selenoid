package config

import (
	"os"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "browsers-*.json")
	assert.NoError(t, err)
	_, err = f.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

func TestNewConfigLoadEmpty(t *testing.T) {
	path := writeTempJSON(t, `{}`)
	conf := NewConfig()
	assert.NoError(t, conf.Load(path, ""))
	assert.NotNil(t, conf.Browsers)
}

func TestCanonicalWebDriverBrowserName(t *testing.T) {
	assert.Equal(t, "MicrosoftEdge", CanonicalWebDriverBrowserName("msedge"))
	assert.Equal(t, "chrome", CanonicalWebDriverBrowserName("chrome"))
}

func TestNormalizeWebDriverBrowserNameInCaps(t *testing.T) {
	caps := map[string]interface{}{"browserName": "msedge"}
	NormalizeWebDriverBrowserNameInCaps(caps)
	assert.Equal(t, "MicrosoftEdge", caps["browserName"])
}

func TestFindPlaywright(t *testing.T) {
	path := writeTempJSON(t, `{
		"playwright-chromium": {
			"default": "1.61.1",
			"versions": {
				"1.61.1": {
					"image": "qaguru/playwright-chromium:1.61.1",
					"port": "4444",
					"path": "/",
					"protocol": "playwright",
					"playwrightVersion": "1.61.1"
				}
			}
		}
	}`)
	conf := NewConfig()
	assert.NoError(t, conf.Load(path, ""))

	browser, version, ok := conf.FindPlaywright("playwright-chromium", "1.61.1")
	assert.True(t, ok)
	assert.Equal(t, "1.61.1", version)
	assert.Equal(t, "playwright", browser.Protocol)
}
