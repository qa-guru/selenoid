package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mafredri/cdp/protocol/network"
	"github.com/stretchr/testify/assert"
)

func headers(raw string) network.Headers {
	return network.Headers([]byte(raw))
}

// feedOneEntry drives a recorder through a full request lifecycle.
func feedOneEntry(r *Recorder, id, url, method string, status int) {
	r.RequestWillBeSent(&network.RequestWillBeSentReply{
		RequestID: network.RequestID(id),
		Timestamp: network.MonotonicTime(1.0),
		WallTime:  network.TimeSinceEpoch(1700000000.25),
		Request: network.Request{
			Method:  method,
			URL:     url,
			Headers: headers(`{"Accept":"*/*","User-Agent":"selenoid-test"}`),
		},
	})
	proto := "http/1.1"
	ip := "93.184.216.34"
	r.ResponseReceived(&network.ResponseReceivedReply{
		RequestID: network.RequestID(id),
		Response: network.Response{
			Status:          status,
			StatusText:      "OK",
			MimeType:        "text/html",
			Headers:         headers(`{"Content-Type":"text/html"}`),
			Protocol:        &proto,
			RemoteIPAddress: &ip,
		},
	})
	r.LoadingFinished(&network.LoadingFinishedReply{
		RequestID:         network.RequestID(id),
		Timestamp:         network.MonotonicTime(1.5),
		EncodedDataLength: 2048,
	})
}

func TestRecorderBuildsValidHAR(t *testing.T) {
	r := NewRecorder()
	feedOneEntry(r, "req-1", "https://example.com/?a=1&b=2", "GET", 200)

	h := r.Build("Example Domain")

	assert.Equal(t, "1.2", h.Log.Version)
	assert.Equal(t, "selenoid", h.Log.Creator.Name)
	assert.Len(t, h.Log.Pages, 1)
	assert.Equal(t, "Example Domain", h.Log.Pages[0].Title)
	assert.Len(t, h.Log.Entries, 1)

	e := h.Log.Entries[0]
	assert.Equal(t, "GET", e.Request.Method)
	assert.Equal(t, "https://example.com/?a=1&b=2", e.Request.URL)
	assert.Equal(t, 200, e.Response.Status)
	assert.Equal(t, "OK", e.Response.StatusText)
	assert.Equal(t, "text/html", e.Response.Content.MimeType)
	assert.Equal(t, "http/1.1", e.Response.HTTPVersion)
	assert.Equal(t, "93.184.216.34", e.ServerIPAddress)
	assert.Equal(t, int64(2048), e.Response.BodySize)
	assert.InDelta(t, 500.0, e.Time, 0.001) // (1.5 - 1.0) * 1000 ms
	assert.Equal(t, "page_1", e.Pageref)

	// Query string is parsed from the URL.
	assert.ElementsMatch(t, []NameValue{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}, e.Request.QueryString)
	// Request headers are surfaced.
	assert.Contains(t, e.Request.Headers, NameValue{Name: "Accept", Value: "*/*"})
}

func TestRecorderMarshalsNonNullArrays(t *testing.T) {
	r := NewRecorder()
	feedOneEntry(r, "req-1", "https://example.com/", "GET", 200)

	data, err := json.Marshal(r.Build(""))
	assert.NoError(t, err)

	// HAR consumers reject null where arrays are required — ensure empties render as [].
	s := string(data)
	assert.NotContains(t, s, `"cookies":null`)
	assert.NotContains(t, s, `"headers":null`)
	assert.NotContains(t, s, `"queryString":null`)
	assert.Contains(t, s, `"cookies":[]`)

	// Round-trips back into the same structure.
	var parsed HAR
	assert.NoError(t, json.Unmarshal(data, &parsed))
	assert.Len(t, parsed.Log.Entries, 1)
}

func TestRecorderHandlesRedirectChain(t *testing.T) {
	r := NewRecorder()
	// First request.
	r.RequestWillBeSent(&network.RequestWillBeSentReply{
		RequestID: network.RequestID("req-1"),
		Timestamp: network.MonotonicTime(1.0),
		WallTime:  network.TimeSinceEpoch(1700000000),
		Request:   network.Request{Method: "GET", URL: "https://example.com/old"},
	})
	// Redirect reuses the same request id and carries the previous response.
	r.RequestWillBeSent(&network.RequestWillBeSentReply{
		RequestID: network.RequestID("req-1"),
		Timestamp: network.MonotonicTime(1.1),
		WallTime:  network.TimeSinceEpoch(1700000001),
		Request:   network.Request{Method: "GET", URL: "https://example.com/new"},
		RedirectResponse: &network.Response{
			Status:     301,
			StatusText: "Moved Permanently",
			Headers:    headers(`{"Location":"https://example.com/new"}`),
		},
	})
	r.LoadingFinished(&network.LoadingFinishedReply{
		RequestID: network.RequestID("req-1"),
		Timestamp: network.MonotonicTime(1.2),
	})

	h := r.Build("")
	assert.Len(t, h.Log.Entries, 2)
	assert.Equal(t, 301, h.Log.Entries[0].Response.Status)
	assert.Equal(t, "https://example.com/new", h.Log.Entries[0].Response.RedirectURL)
	assert.Equal(t, "https://example.com/new", h.Log.Entries[1].Request.URL)
}

func TestRecorderIncludesPendingRequests(t *testing.T) {
	r := NewRecorder()
	// Request started but never finished (session ended mid-flight).
	r.RequestWillBeSent(&network.RequestWillBeSentReply{
		RequestID: network.RequestID("req-1"),
		Timestamp: network.MonotonicTime(1.0),
		WallTime:  network.TimeSinceEpoch(1700000000),
		Request:   network.Request{Method: "GET", URL: "https://example.com/hang"},
	})
	assert.Equal(t, 1, r.EntryCount())
	h := r.Build("")
	assert.Len(t, h.Log.Entries, 1)
	assert.Equal(t, "https://example.com/hang", h.Log.Entries[0].Request.URL)
}

func TestRecorderPostData(t *testing.T) {
	r := NewRecorder()
	body := `{"hello":"world"}`
	r.RequestWillBeSent(&network.RequestWillBeSentReply{
		RequestID: network.RequestID("req-1"),
		Timestamp: network.MonotonicTime(1.0),
		WallTime:  network.TimeSinceEpoch(1700000000),
		Request: network.Request{
			Method:   "POST",
			URL:      "https://example.com/api",
			Headers:  headers(`{"Content-Type":"application/json"}`),
			PostData: &body,
		},
	})
	h := r.Build("")
	assert.Len(t, h.Log.Entries, 1)
	pd := h.Log.Entries[0].Request.PostData
	assert.NotNil(t, pd)
	assert.Equal(t, "application/json", pd.MimeType)
	assert.Equal(t, body, pd.Text)
	assert.Equal(t, int64(len(body)), h.Log.Entries[0].Request.BodySize)
}

func TestRecorderWriteFile(t *testing.T) {
	r := NewRecorder()
	feedOneEntry(r, "req-1", "https://example.com/", "GET", 200)

	dir := t.TempDir()
	path := filepath.Join(dir, "session.har")
	assert.NoError(t, r.WriteFile(path, "title"))

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	var parsed HAR
	assert.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "1.2", parsed.Log.Version)
	assert.Len(t, parsed.Log.Entries, 1)
}
