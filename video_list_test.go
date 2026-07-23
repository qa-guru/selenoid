package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVideoListLimitOffset(t *testing.T) {
	t.Run("defaults to page size 10", func(t *testing.T) {
		limit, offset := parseVideoListLimitOffset(url.Values{})
		assert.Equal(t, 10, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("clamps limit to max 100", func(t *testing.T) {
		limit, _ := parseVideoListLimitOffset(url.Values{"limit": []string{"500"}})
		assert.Equal(t, 100, limit)
	})

	t.Run("rejects non-positive limit", func(t *testing.T) {
		limit, _ := parseVideoListLimitOffset(url.Values{"limit": []string{"0"}})
		assert.Equal(t, 10, limit)
	})

	t.Run("rejects negative offset", func(t *testing.T) {
		_, offset := parseVideoListLimitOffset(url.Values{"offset": []string{"-5"}})
		assert.Equal(t, 0, offset)
	})
}

func TestPaginateFileNames(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

	t.Run("first page of 10", func(t *testing.T) {
		page := paginateFileNames(names, 10, 0)
		assert.Equal(t, names[:10], page)
	})

	t.Run("second page", func(t *testing.T) {
		page := paginateFileNames(names, 10, 10)
		assert.Equal(t, []string{"k", "l"}, page)
	})

	t.Run("offset past end", func(t *testing.T) {
		page := paginateFileNames(names, 10, 100)
		assert.Empty(t, page)
	})
}

func TestFilterFileNames(t *testing.T) {
	names := []string{"abc.mp4", "def.mp4", "abc-extra.mp4"}
	assert.Equal(t, []string{"abc.mp4", "abc-extra.mp4"}, filterFileNames(names, "abc"))
	assert.Equal(t, names, filterFileNames(names, ""))
}

func TestListVideosAsJsonPagination(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 25; i++ {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("video-%02d.mp4", i)), []byte("x"), 0644))
	}

	req := httptest.NewRequest(http.MethodGet, "/video/?json&limit=10&offset=10", nil)
	rr := httptest.NewRecorder()
	listVideosAsJson(1, rr, req, dir)
	assert.Equal(t, http.StatusOK, rr.Code)

	var listed videoListResponse
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	assert.Equal(t, 25, listed.Total)
	assert.Equal(t, 10, listed.Limit)
	assert.Equal(t, 10, listed.Offset)
	assert.Len(t, listed.Videos, 10)

	req = httptest.NewRequest(http.MethodGet, "/video/?json", nil)
	rr = httptest.NewRecorder()
	listVideosAsJson(1, rr, req, dir)
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	assert.Equal(t, 25, listed.Total)
	assert.LessOrEqual(t, len(listed.Videos), 10)
	assert.Equal(t, defaultVideoListLimit, listed.Limit)

	req = httptest.NewRequest(http.MethodGet, "/video/?json&q=video-01", nil)
	rr = httptest.NewRecorder()
	listVideosAsJson(1, rr, req, dir)
	assert.NoError(t, json.NewDecoder(rr.Body).Decode(&listed))
	assert.Equal(t, 1, listed.Total)
	assert.Equal(t, []string{"video-01.mp4"}, listed.Videos)
}
