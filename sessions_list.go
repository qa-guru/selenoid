package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/qa-guru/selenoid/info"
	"github.com/qa-guru/selenoid/session"
)

// sessionArtifacts is one finished session's artifacts, grouped by session id
// (the shared base name of the <id>.mp4 / <id>.log / <id>.har files). Fields are
// the concrete file names so the UI can link straight to /video/<video>,
// /logs/<log> and /har/<har> (and DELETE each of them for a whole-session wipe).
// Optional name/quota/started/finished come from the sidecar metadata JSON
// written when the session stops (and finished falls back to newest artifact mtime).
type sessionArtifacts struct {
	ID       string     `json:"id"`
	Video    string     `json:"video,omitempty"`
	Log      string     `json:"log,omitempty"`
	HAR      string     `json:"har,omitempty"`
	Name     string     `json:"name,omitempty"`
	Quota    string     `json:"quota,omitempty"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
}

type sessionListResponse struct {
	Sessions []sessionArtifacts `json:"sessions"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// sessionsList serves a session-centric view over the video/log/har artifact
// directories, mirroring the /video/ and /har/ JSON listing contract:
//
//	GET /sessions/?json[&limit=&offset=&q=]
//
// Files are grouped by the base name shared by <id>.mp4, <id>.log and <id>.har
// (the session id for default naming). Grouping only lines up when the three
// artifacts share a base name; custom videoName/harName values that differ from
// the session id surface as separate entries. Deletion is intentionally not
// handled here: the UI performs a whole-session wipe by issuing the existing
// per-artifact DELETE /video/<f>, DELETE /logs/<f>, DELETE /har/<f> requests.
func sessionsList(w http.ResponseWriter, r *http.Request) {
	requestId := serial()
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, remote := info.RequestInfo(r)
	log.Printf("[%d] [SESSIONS_LISTING] [%s] [%s]", requestId, user, remote)

	group := collectSessionArtifacts()
	ids := make([]string, 0, len(group))
	for id := range group {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	query := r.URL.Query()
	ids = filterFileNames(ids, query.Get("q"))
	limit, offset := parseVideoListLimitOffset(query)
	page := paginateFileNames(ids, limit, offset)

	out := make([]sessionArtifacts, 0, len(page))
	for _, id := range page {
		out = append(out, *group[id])
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessionListResponse{
		Sessions: out,
		Total:    len(ids),
		Limit:    limit,
		Offset:   offset,
	})
}

// collectSessionArtifacts scans the configured artifact directories and groups
// files by their base name (extension stripped). Missing/unconfigured
// directories are skipped so the endpoint degrades gracefully when logs or HAR
// are disabled. Sidecar <id>.json metadata (when present) fills name/quota/times.
func collectSessionArtifacts() map[string]*sessionArtifacts {
	group := map[string]*sessionArtifacts{}
	touch := func(id string) *sessionArtifacts {
		a := group[id]
		if a == nil {
			a = &sessionArtifacts{ID: id}
			group[id] = a
		}
		return a
	}
	add := func(dir, ext string, set func(a *sessionArtifacts, name string)) {
		if dir == "" {
			return
		}
		files, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, f := range files {
			name := f.Name()
			if f.IsDir() || !strings.HasSuffix(name, ext) {
				continue
			}
			id := strings.TrimSuffix(name, ext)
			a := touch(id)
			set(a, name)
			if info, err := f.Info(); err == nil {
				mt := info.ModTime()
				if a.Finished == nil || mt.After(*a.Finished) {
					t := mt
					a.Finished = &t
				}
			}
		}
	}
	add(videoOutputDir, videoFileExtension, func(a *sessionArtifacts, name string) { a.Video = name })
	add(logOutputDir, logFileExtension, func(a *sessionArtifacts, name string) { a.Log = name })
	add(harOutputDir, harFileExtension, func(a *sessionArtifacts, name string) { a.HAR = name })

	for _, a := range group {
		applySessionMetadata(a)
	}
	return group
}

func applySessionMetadata(a *sessionArtifacts) {
	if a == nil || logOutputDir == "" {
		return
	}
	path := filepath.Join(logOutputDir, a.ID+metadataFileExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var meta session.Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return
	}
	if meta.Capabilities.TestName != "" {
		a.Name = meta.Capabilities.TestName
	}
	if meta.Quota != "" {
		a.Quota = meta.Quota
	}
	if !meta.Started.IsZero() {
		t := meta.Started
		a.Started = &t
	}
	if !meta.Finished.IsZero() {
		t := meta.Finished
		a.Finished = &t
	}
}
