package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aerokube/selenoid/config"
	"github.com/aerokube/selenoid/session"
	assert "github.com/stretchr/testify/require"
)

const testLogConf = "config/container-logs.json"

func configfile(s string) string {
	tmp, err := os.CreateTemp("", "config")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tmp.Write([]byte(s))
	if err != nil {
		log.Fatal(err)
	}
	err = tmp.Close()
	if err != nil {
		log.Fatal(err)
	}
	return tmp.Name()
}

func TestConfig(t *testing.T) {
	t.Run("Config", func(t *testing.T) {
		confFile := configfile(`{}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)
	})
}

func TestConfigError(t *testing.T) {
	t.Run("Config error", func(t *testing.T) {
		confFile := configfile(`{}`)
		_ = os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), fmt.Sprintf("browsers config: read error: open %s: no such file or directory", confFile))
	})
}

func TestLogConfigError(t *testing.T) {
	t.Run("Log config error", func(t *testing.T) {
		confFile := configfile(`{}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, "some-missing-file")
		assert.Error(t, err)
	})
}

func TestConfigParseError(t *testing.T) {
	t.Run("Config parse error", func(t *testing.T) {
		confFile := configfile(`{`)
		defer os.Remove(confFile)
		var conf config.Config
		err := conf.Load(confFile, testLogConf)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "browsers config: parse error: unexpected end of JSON input")
	})
}

func TestConfigEmptyState(t *testing.T) {
	t.Run("Config empty state", func(t *testing.T) {
		confFile := configfile(`{}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		conf.Load(confFile, testLogConf)

		state := conf.State(session.NewMap(), 0, 0, 0)
		assert.Equal(t, state.Total, 0)
		assert.Equal(t, state.Queued, 0)
		assert.Equal(t, state.Pending, 0)
		assert.Equal(t, state.Used, 0)
	})
}

func TestConfigNonEmptyState(t *testing.T) {
	t.Run("Config non empty state", func(t *testing.T) {
		confFile := configfile(`{}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		_ = conf.Load(confFile, testLogConf)

		sessions := session.NewMap()
		sessions.Put("0", &session.Session{Caps: session.Caps{Name: "firefox", Version: "49.0"}, Quota: "unknown"})
		state := conf.State(sessions, 1, 0, 0)
		assert.Equal(t, state.Total, 1)
		assert.Equal(t, state.Queued, 0)
		assert.Equal(t, state.Pending, 0)
		assert.Equal(t, state.Used, 1)
		assert.Equal(t, state.Browsers["firefox"]["49.0"]["unknown"].Count, 1)
	})
}

func TestConfigEmptyVersions(t *testing.T) {
	t.Run("Config empty versions", func(t *testing.T) {
		confFile := configfile(`{"firefox":{}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		_ = conf.Load(confFile, testLogConf)

		sessions := session.NewMap()
		sessions.Put("0", &session.Session{Caps: session.Caps{Name: "firefox", Version: "49.0"}, Quota: "unknown"})
		state := conf.State(sessions, 1, 0, 0)
		assert.Equal(t, state.Total, 1)
		assert.Equal(t, state.Queued, 0)
		assert.Equal(t, state.Pending, 0)
		assert.Equal(t, state.Used, 1)
		assert.Equal(t, state.Browsers["firefox"]["49.0"]["unknown"].Count, 1)
	})
}

func TestConfigNonEmptyVersions(t *testing.T) {
	t.Run("Config non empty versions", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		conf.Load(confFile, testLogConf)

		sessions := session.NewMap()
		sessions.Put("0", &session.Session{Caps: session.Caps{Name: "firefox", Version: "49.0"}, Quota: "unknown"})
		state := conf.State(sessions, 1, 0, 0)
		assert.Equal(t, state.Total, 1)
		assert.Equal(t, state.Queued, 0)
		assert.Equal(t, state.Pending, 0)
		assert.Equal(t, state.Used, 1)
		assert.Equal(t, state.Browsers["firefox"]["49.0"]["unknown"].Count, 1)
	})
}

func TestConfigFindMissingBrowser(t *testing.T) {
	t.Run("Config find missing browser", func(t *testing.T) {
		confFile := configfile(`{}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		conf.Load(confFile, testLogConf)

		_, _, ok := conf.Find("firefox", "")
		assert.False(t, ok)
	})
}

func TestConfigFindDefaultVersionError(t *testing.T) {
	t.Run("Config find default version error", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":""}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		conf.Load(confFile, testLogConf)

		_, _, ok := conf.Find("firefox", "")
		assert.False(t, ok)
	})
}

func TestConfigFindDefaultVersion(t *testing.T) {
	t.Run("Config find default version", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0"}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		_, v, ok := conf.Find("firefox", "")
		assert.False(t, ok)
		assert.Equal(t, v, "49.0")
	})
}

func TestConfigFindFoundByEmptyPrefix(t *testing.T) {
	t.Run("Config find found by empty prefix", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		_, v, ok := conf.Find("firefox", "")
		assert.True(t, ok)
		assert.Equal(t, v, "49.0")
	})
}

func TestConfigFindFoundByPrefix(t *testing.T) {
	t.Run("Config find found by prefix", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		_, v, ok := conf.Find("firefox", "49")
		assert.True(t, ok)
		assert.Equal(t, v, "49.0")
	})
}

func TestConfigFindFoundByMatch(t *testing.T) {
	t.Run("Config find found by match", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		_, v, ok := conf.Find("firefox", "49.0")
		assert.True(t, ok)
		assert.Equal(t, v, "49.0")
	})
}

func TestConfigFindImage(t *testing.T) {
	t.Run("Config find image", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{"image":"image","port":"5555", "path":"/"}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		b, v, ok := conf.Find("firefox", "49.0")
		assert.True(t, ok)
		assert.Equal(t, v, "49.0")
		assert.Equal(t, b.Image, "image")
		assert.Equal(t, b.Port, "5555")
		assert.Equal(t, b.Path, "/")
	})
}

func TestConfigFindMicrosoftEdgeAlias(t *testing.T) {
	t.Run("Config find microsoft edge alias", func(t *testing.T) {
		confFile := configfile(`{"msedge":{"default":"145.0","versions":{"145.0":{"image":"edge","port":"4444","path":"/"}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)

		b, v, ok := conf.Find("MicrosoftEdge", "145.0")
		assert.True(t, ok)
		assert.Equal(t, v, "145.0")
		assert.Equal(t, b.Image, "edge")
	})
}

func TestConfigConcurrentLoad(t *testing.T) {
	t.Run("Config concurrent load", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":""}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()

		done := make(chan struct{})
		go func() {
			_ = conf.Load(confFile, testLogConf)
			done <- struct{}{}
		}()
		_ = conf.Load(confFile, testLogConf)
		<-done
	})
}

func TestConfigConcurrentLoadAndRead(t *testing.T) {
	t.Run("Config concurrent load and read", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{"image":"image","port":"5555", "path":"/", "tmpfs": {"/tmp":"size=64k"}}}}}`)
		defer os.Remove(confFile)
		conf := config.NewConfig()
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)
		done := make(chan string)
		go func() {
			browser, _, _ := conf.Find("firefox", "")
			done <- browser.Tmpfs["/tmp"]
		}()
		err = conf.Load(confFile, testLogConf)
		assert.NoError(t, err)
		<-done
	})
}

func TestConfigConcurrentRead(t *testing.T) {
	t.Run("Config concurrent read", func(t *testing.T) {
		confFile := configfile(`{"firefox":{"default":"49.0","versions":{"49.0":{"image":"image","port":"5555", "path":"/", "tmpfs": {"/tmp":"size=64k"}}}}}`)
		defer os.Remove(confFile)
		var conf config.Config
		err := conf.Load(confFile, testLogConf)
		assert.NoError(t, err)
		done := make(chan string)
		go func() {
			browser, _, _ := conf.Find("firefox", "")
			done <- browser.Tmpfs["/tmp"]
		}()
		go func() {
			browser, _, _ := conf.Find("firefox", "")
			done <- browser.Tmpfs["/tmp"]
		}()
		<-done
		<-done
	})
}
