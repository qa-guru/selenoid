package session

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCapsProcessExtensionCapabilities(t *testing.T) {
	caps := Caps{
		W3CVersion:  "120.0",
		W3CPlatform: "linux",
		W3CDeviceName: "pixel",
		ExtensionCapabilities: &Caps{
			VNC:   true,
			Video: true,
			Name:  "chrome",
		},
	}
	caps.ProcessExtensionCapabilities()
	assert.Equal(t, "120.0", caps.Version)
	assert.Equal(t, "linux", caps.Platform)
	assert.Equal(t, "pixel", caps.DeviceName)
	assert.True(t, caps.VNC)
	assert.True(t, caps.Video)
	assert.Equal(t, "chrome", caps.Name)
}

func TestCapsBrowserName(t *testing.T) {
	assert.Equal(t, "firefox", (&Caps{Name: "firefox"}).BrowserName())
	assert.Equal(t, "iphone", (&Caps{DeviceName: "iphone"}).BrowserName())
	assert.Equal(t, "android", (&Caps{W3CDeviceName: "android"}).BrowserName())
	assert.Equal(t, "", (&Caps{}).BrowserName())
}

func TestMapPutGetRemoveEachLen(t *testing.T) {
	m := NewMap()
	s := &Session{Quota: "user1"}
	m.Put("abc", s)
	got, ok := m.Get("abc")
	assert.True(t, ok)
	assert.Equal(t, s, got)
	assert.Equal(t, 1, m.Len())

	seen := 0
	m.Each(func(k string, v *Session) {
		assert.Equal(t, "abc", k)
		assert.Equal(t, s, v)
		seen++
	})
	assert.Equal(t, 1, seen)

	m.Remove("abc")
	_, ok = m.Get("abc")
	assert.False(t, ok)
	assert.Equal(t, 0, m.Len())
}
