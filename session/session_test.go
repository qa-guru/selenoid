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

func TestCapsHARContentNormalizeAndMerge(t *testing.T) {
	meta := Caps{}
	meta.NormalizeHARContent()
	assert.Equal(t, "", meta.HARContent)
	assert.False(t, meta.HARBodies())

	bodies := Caps{HARContent: "BODIES"}
	bodies.NormalizeHARContent()
	assert.Equal(t, "bodies", bodies.HARContent)
	assert.True(t, bodies.HARBodies())

	explicitMeta := Caps{HARContent: "meta"}
	explicitMeta.NormalizeHARContent()
	assert.Equal(t, "meta", explicitMeta.HARContent)
	assert.False(t, explicitMeta.HARBodies())

	unknown := Caps{HARContent: "full"}
	unknown.NormalizeHARContent()
	assert.Equal(t, "", unknown.HARContent)
	assert.False(t, unknown.HARBodies())

	// selenoid:options merge carries harContent.
	caps := Caps{
		ExtensionCapabilities: &Caps{
			HAR:        true,
			HARContent: "bodies",
		},
	}
	caps.ProcessExtensionCapabilities()
	assert.True(t, caps.HAR)
	assert.Equal(t, "bodies", caps.HARContent)
	assert.True(t, caps.HARBodies())
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
