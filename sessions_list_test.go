package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qa-guru/selenoid/session"
	"github.com/stretchr/testify/assert"
)

// setArtifactDirs points the package-level video/log/har output dirs at temp
// dirs for the duration of a test and restores the previous values afterwards.
func setArtifactDirs(t *testing.T, video, logs, har string) {
	t.Helper()
	pv, pl, ph := videoOutputDir, logOutputDir, harOutputDir
	videoOutputDir, logOutputDir, harOutputDir = video, logs, har
	t.Cleanup(func() { videoOutputDir, logOutputDir, harOutputDir = pv, pl, ph })
}

func decodeSessions(t *testing.T, rr *httptest.ResponseRecorder) sessionListResponse {
	t.Helper()
	var listed sessionListResponse
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	return listed
}

func TestSessionsListGroupsArtifactsById(t *testing.T) {
	video := t.TempDir()
	logs := t.TempDir()
	har := t.TempDir()
	setArtifactDirs(t, video, logs, har)

	// session-a has all three, session-b only video+log, session-c only har.
	assert.NoError(t, os.WriteFile(filepath.Join(video, "session-a.mp4"), []byte("v"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(logs, "session-a.log"), []byte("l"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(har, "session-a.har"), []byte("{}"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(video, "session-b.mp4"), []byte("v"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(logs, "session-b.log"), []byte("l"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(har, "session-c.har"), []byte("{}"), 0644))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.Bytes()
	listed := decodeSessions(t, rr)
	assert.Equal(t, 3, listed.Total)
	assert.Len(t, listed.Sessions, 3)

	byID := map[string]sessionArtifacts{}
	for _, s := range listed.Sessions {
		byID[s.ID] = s
	}
	assert.Equal(t, "session-a.mp4", byID["session-a"].Video)
	assert.Equal(t, "session-a.log", byID["session-a"].Log)
	assert.Equal(t, "session-a.har", byID["session-a"].HAR)
	assert.NotNil(t, byID["session-a"].Finished, "finished falls back to artifact mtime")
	assert.Nil(t, byID["session-a"].Started, "started omitted without metadata sidecar")
	assert.Equal(t, "session-b.mp4", byID["session-b"].Video)
	assert.Equal(t, "session-b.log", byID["session-b"].Log)
	assert.Equal(t, "session-c.har", byID["session-c"].HAR)

	// Zero time.Time must not leak as 0001-01-01 in JSON (UI would show junk).
	var raw map[string]any
	assert.NoError(t, json.Unmarshal(body, &raw))
	sessions, ok := raw["sessions"].([]any)
	assert.True(t, ok)
	for _, item := range sessions {
		m := item.(map[string]any)
		assert.NotContains(t, m, "started")
		assert.Contains(t, m, "finished")
	}
}

func TestSessionsListLinksHarFromMetadataHarName(t *testing.T) {
	video := t.TempDir()
	logs := t.TempDir()
	har := t.TempDir()
	setArtifactDirs(t, video, logs, har)

	assert.NoError(t, os.WriteFile(filepath.Join(video, "sess-har.mp4"), []byte("v"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(har, "custom-name.har"), []byte("{}"), 0644))
	meta := session.Metadata{
		ID:    "sess-har",
		Quota: "user1",
		Capabilities: session.Caps{
			HAR:     true,
			HARName: "custom-name.har",
		},
	}
	raw, err := json.MarshalIndent(meta, "", "    ")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(logs, "sess-har.json"), raw, 0644))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	byID := map[string]sessionArtifacts{}
	for _, s := range listed.Sessions {
		byID[s.ID] = s
	}
	got, ok := byID["sess-har"]
	assert.True(t, ok)
	assert.Equal(t, "sess-har.mp4", got.Video)
	assert.Equal(t, "custom-name.har", got.HAR)
}

func TestSessionsListEnrichesFromMetadata(t *testing.T) {
	video := t.TempDir()
	logs := t.TempDir()
	setArtifactDirs(t, video, logs, "")

	assert.NoError(t, os.WriteFile(filepath.Join(video, "sess-meta.mp4"), []byte("v"), 0644))
	started := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	finished := started.Add(95 * time.Second)
	meta := session.Metadata{
		ID:       "sess-meta",
		Quota:    "alice",
		Started:  started,
		Finished: finished,
		Capabilities: session.Caps{
			TestName: "MyCoolTest.shouldPass",
		},
	}
	raw, err := json.MarshalIndent(meta, "", "    ")
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(logs, "sess-meta.json"), raw, 0644))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	assert.Len(t, listed.Sessions, 1)
	got := listed.Sessions[0]
	assert.Equal(t, "sess-meta", got.ID)
	assert.Equal(t, "sess-meta.mp4", got.Video)
	assert.Equal(t, "alice", got.Quota)
	assert.Equal(t, "MyCoolTest.shouldPass", got.Name)
	if assert.NotNil(t, got.Started) {
		assert.True(t, got.Started.Equal(started))
	}
	if assert.NotNil(t, got.Finished) {
		assert.True(t, got.Finished.Equal(finished))
	}
}

func TestSessionsListPaginationAndSort(t *testing.T) {
	video := t.TempDir()
	setArtifactDirs(t, video, "", "")
	for i := 0; i < 15; i++ {
		assert.NoError(t, os.WriteFile(filepath.Join(video, fmt.Sprintf("session-%02d.mp4", i)), []byte("v"), 0644))
	}

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json&limit=10&offset=10&sort=id&order=asc", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	assert.Equal(t, 15, listed.Total)
	assert.Equal(t, 10, listed.Limit)
	assert.Equal(t, 10, listed.Offset)
	assert.Len(t, listed.Sessions, 5)
	// Sorted ascending by id -> offset 10 is session-10.
	assert.Equal(t, "session-10", listed.Sessions[0].ID)
	assert.Equal(t, "session-14", listed.Sessions[4].ID)
}

func TestSessionsListDefaultSortNewestFinishedFirst(t *testing.T) {
	video := t.TempDir()
	logs := t.TempDir()
	setArtifactDirs(t, video, logs, "")

	writeMeta := func(id string, finished time.Time) {
		meta := session.Metadata{
			ID:       id,
			Finished: finished,
		}
		raw, err := json.Marshal(meta)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(filepath.Join(logs, id+".json"), raw, 0644))
		assert.NoError(t, os.WriteFile(filepath.Join(video, id+".mp4"), []byte("v"), 0644))
	}

	writeMeta("old", time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC))
	writeMeta("new", time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC))
	writeMeta("mid", time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	assert.Equal(t, []string{"new", "mid", "old"}, []string{
		listed.Sessions[0].ID,
		listed.Sessions[1].ID,
		listed.Sessions[2].ID,
	})
}

func TestSessionsListSortByDurationDesc(t *testing.T) {
	video := t.TempDir()
	logs := t.TempDir()
	setArtifactDirs(t, video, logs, "")

	writeMeta := func(id string, started, finished time.Time) {
		meta := session.Metadata{
			ID:       id,
			Started:  started,
			Finished: finished,
		}
		raw, err := json.Marshal(meta)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(filepath.Join(logs, id+".json"), raw, 0644))
		assert.NoError(t, os.WriteFile(filepath.Join(video, id+".mp4"), []byte("v"), 0644))
	}

	base := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	writeMeta("short", base, base.Add(5*time.Second))
	writeMeta("long", base, base.Add(2*time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json&sort=duration&order=desc", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	assert.Equal(t, []string{"long", "short"}, []string{
		listed.Sessions[0].ID,
		listed.Sessions[1].ID,
	})
}

func TestSessionsListFilter(t *testing.T) {
	video := t.TempDir()
	har := t.TempDir()
	setArtifactDirs(t, video, "", har)
	assert.NoError(t, os.WriteFile(filepath.Join(video, "alpha.mp4"), []byte("v"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(har, "beta.har"), []byte("{}"), 0644))

	req := httptest.NewRequest(http.MethodGet, "/sessions/?json&q=alpha", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	listed := decodeSessions(t, rr)
	assert.Equal(t, 1, listed.Total)
	assert.Len(t, listed.Sessions, 1)
	assert.Equal(t, "alpha", listed.Sessions[0].ID)
}

func TestSessionsListEmptyWhenNoDirs(t *testing.T) {
	setArtifactDirs(t, "", "", "")
	req := httptest.NewRequest(http.MethodGet, "/sessions/?json", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()
	// Non-null empty array in JSON.
	assert.Contains(t, body, `"sessions":[]`)

	var listed sessionListResponse
	assert.NoError(t, json.Unmarshal([]byte(body), &listed))
	assert.Equal(t, 0, listed.Total)
}

func TestSessionsListMethodNotAllowed(t *testing.T) {
	setArtifactDirs(t, t.TempDir(), "", "")
	req := httptest.NewRequest(http.MethodDelete, "/sessions/session-a", nil)
	rr := httptest.NewRecorder()
	sessionsList(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
