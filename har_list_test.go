package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/qa-guru/selenoid/session"
	"github.com/stretchr/testify/assert"
)

// setHarOutputDir points the package-level harOutputDir at a temp dir for the
// duration of a test and restores the previous value afterwards.
func setHarOutputDir(t *testing.T, dir string) {
	t.Helper()
	prev := harOutputDir
	harOutputDir = dir
	t.Cleanup(func() { harOutputDir = prev })
}

func TestHarDisabledReturns404(t *testing.T) {
	setHarOutputDir(t, "")
	req := httptest.NewRequest(http.MethodGet, "/har/?json", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHarListJson(t *testing.T) {
	dir := t.TempDir()
	setHarOutputDir(t, dir)
	for i := 0; i < 15; i++ {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("session-%02d.har", i)), []byte("{}"), 0644))
	}

	req := httptest.NewRequest(http.MethodGet, "/har/?json&limit=10&offset=10", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var listed videoListResponse
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	assert.Equal(t, 15, listed.Total)
	assert.Equal(t, 10, listed.Limit)
	assert.Equal(t, 10, listed.Offset)
	assert.Len(t, listed.Videos, 5)
}

func TestHarListJsonFilter(t *testing.T) {
	dir := t.TempDir()
	setHarOutputDir(t, dir)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "alpha.har"), []byte("{}"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "beta.har"), []byte("{}"), 0644))

	req := httptest.NewRequest(http.MethodGet, "/har/?json&q=alpha", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var listed videoListResponse
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	assert.Equal(t, 1, listed.Total)
	assert.Equal(t, []string{"alpha.har"}, listed.Videos)
}

func TestHarDownload(t *testing.T) {
	dir := t.TempDir()
	setHarOutputDir(t, dir)
	content := `{"log":{"version":"1.2"}}`
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "session.har"), []byte(content), 0644))

	req := httptest.NewRequest(http.MethodGet, "/har/session.har", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, content, rr.Body.String())
}

func TestHarDelete(t *testing.T) {
	dir := t.TempDir()
	setHarOutputDir(t, dir)
	path := filepath.Join(dir, "session.har")
	assert.NoError(t, os.WriteFile(path, []byte("{}"), 0644))

	req := httptest.NewRequest(http.MethodDelete, "/har/session.har", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestHarDeleteMissing(t *testing.T) {
	dir := t.TempDir()
	setHarOutputDir(t, dir)
	req := httptest.NewRequest(http.MethodDelete, "/har/nope.har", nil)
	rr := httptest.NewRecorder()
	har(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHarCaptureEnabled(t *testing.T) {
	setHarOutputDir(t, t.TempDir())
	prevDisable := disableDocker
	disableDocker = false
	t.Cleanup(func() { disableDocker = prevDisable })

	// Enabled: har cap set, docker on, devtools endpoint present.
	assert.True(t, harCaptureEnabled(session.Caps{HAR: true}, "172.17.0.2:7070"))
	// Disabled: cap off.
	assert.False(t, harCaptureEnabled(session.Caps{HAR: false}, "172.17.0.2:7070"))
	// Disabled: no devtools endpoint (e.g. non-docker or Playwright server).
	assert.False(t, harCaptureEnabled(session.Caps{HAR: true}, ""))

	// Disabled: har-output-dir not configured.
	setHarOutputDir(t, "")
	assert.False(t, harCaptureEnabled(session.Caps{HAR: true}, "172.17.0.2:7070"))
}
