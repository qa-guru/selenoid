// Package har builds HTTP Archive (HAR 1.2) files for Selenoid sessions.
//
// The hub records browser network activity over the Chrome DevTools Protocol
// (CDP) — the same endpoint that powers the `se:cdp` capability and the
// `/devtools/<session-id>/` proxy (browser container port 7070). This is the
// native, browser-side capture path: unlike video (ffmpeg) it can actually
// see the network, and unlike a MITM sidecar it needs no extra container.
//
// The package is split into two layers so the HAR assembly is unit-testable
// without a live browser:
//
//   - Recorder — a pure event sink. Feed it CDP network.*Reply structs and it
//     builds a valid HAR. No I/O, no goroutines.
//   - Session — the live CDP connection. It dials the DevTools websocket,
//     enables the Network domain and streams events into a Recorder.
package har

import (
	"encoding/json"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mafredri/cdp/protocol/network"
)

// Creator identifies the tool that produced the HAR.
const (
	creatorName = "selenoid"
	harVersion  = "1.2"
)

// HAR is the root of an HTTP Archive document (HAR 1.2).
type HAR struct {
	Log Log `json:"log"`
}

// Log holds the recorded pages and entries.
type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Pages   []Page  `json:"pages"`
	Entries []Entry `json:"entries"`
}

// Creator describes the application that created the log.
type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Page groups entries that belong to a single navigation.
type Page struct {
	StartedDateTime string      `json:"startedDateTime"`
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	PageTimings     PageTimings `json:"pageTimings"`
}

// PageTimings holds page-level load milestones (ms).
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad"`
	OnLoad        float64 `json:"onLoad"`
}

// Entry is a single request/response pair.
type Entry struct {
	Pageref         string   `json:"pageref,omitempty"`
	StartedDateTime string   `json:"startedDateTime"`
	Time            float64  `json:"time"`
	Request         Request  `json:"request"`
	Response        Response `json:"response"`
	Cache           Cache    `json:"cache"`
	Timings         Timings  `json:"timings"`
	ServerIPAddress string   `json:"serverIPAddress,omitempty"`
	Connection      string   `json:"connection,omitempty"`
}

// Request captures the outgoing HTTP request.
type Request struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []Cookie    `json:"cookies"`
	Headers     []NameValue `json:"headers"`
	QueryString []NameValue `json:"queryString"`
	PostData    *PostData   `json:"postData,omitempty"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

// Response captures the incoming HTTP response.
type Response struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []Cookie    `json:"cookies"`
	Headers     []NameValue `json:"headers"`
	Content     Content     `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

// Content describes the response body (text omitted by default).
type Content struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// Cache is always empty — Selenoid does not track cache state.
type Cache struct{}

// Timings breaks the entry time into phases (ms). Unknown phases are -1 per spec.
type Timings struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// Cookie is a single request/response cookie (name/value only).
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NameValue is a generic header or query-string pair.
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PostData holds request body text for non-GET requests.
type PostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// Recorder turns CDP network events into HAR entries. It is safe for
// concurrent use so multiple CDP event streams can feed it in parallel.
type Recorder struct {
	mu        sync.Mutex
	pending   map[network.RequestID]*entryState
	completed []*entryState
	seq       int64
}

// entryState is an in-progress entry plus the bookkeeping needed to finalize it.
type entryState struct {
	seq       int64
	entry     Entry
	startMono float64
	finished  bool
}

// NewRecorder returns an empty recorder.
func NewRecorder() *Recorder {
	return &Recorder{pending: make(map[network.RequestID]*entryState)}
}

// RequestWillBeSent records the start of a request. Redirects reuse the same
// request id, so a redirectResponse finalizes the previous entry first.
func (r *Recorder) RequestWillBeSent(ev *network.RequestWillBeSentReply) {
	if ev == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if ev.RedirectResponse != nil {
		if prev, ok := r.pending[ev.RequestID]; ok {
			applyResponse(&prev.entry, ev.RedirectResponse)
			prev.entry.Response.RedirectURL = redirectLocation(ev.RedirectResponse)
			r.complete(ev.RequestID, prev)
		}
	}

	req := ev.Request
	e := &entryState{
		seq:       r.seq,
		startMono: float64(ev.Timestamp),
		entry: Entry{
			StartedDateTime: epochToRFC3339(float64(ev.WallTime)),
			Time:            0,
			Request: Request{
				Method:      req.Method,
				URL:         req.URL,
				HTTPVersion: "HTTP/1.1",
				Cookies:     []Cookie{},
				Headers:     decodeHeaders(req.Headers),
				QueryString: queryString(req.URL),
				PostData:    postData(&req),
				HeadersSize: -1,
				BodySize:    bodySize(req.PostData),
			},
			Response: Response{
				Status:      0,
				StatusText:  "",
				HTTPVersion: "HTTP/1.1",
				Cookies:     []Cookie{},
				Headers:     []NameValue{},
				Content:     Content{Size: 0, MimeType: ""},
				RedirectURL: "",
				HeadersSize: -1,
				BodySize:    -1,
			},
			Cache:   Cache{},
			Timings: Timings{Send: 0, Wait: -1, Receive: -1},
		},
	}
	r.seq++
	r.pending[ev.RequestID] = e
}

// ResponseReceived attaches response metadata to a pending entry.
func (r *Recorder) ResponseReceived(ev *network.ResponseReceivedReply) {
	if ev == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.pending[ev.RequestID]; ok {
		applyResponse(&e.entry, &ev.Response)
	}
}

// LoadingFinished finalizes a pending entry once the body has fully loaded.
func (r *Recorder) LoadingFinished(ev *network.LoadingFinishedReply) {
	if ev == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.pending[ev.RequestID]; ok {
		if ev.EncodedDataLength > 0 {
			e.entry.Response.BodySize = int64(ev.EncodedDataLength)
		}
		e.entry.Time = durationMs(e.startMono, float64(ev.Timestamp))
		e.entry.Timings.Wait = e.entry.Time
		e.entry.Timings.Receive = 0
		r.complete(ev.RequestID, e)
	}
}

// LoadingFailed finalizes a pending entry that errored or was canceled.
func (r *Recorder) LoadingFailed(ev *network.LoadingFailedReply) {
	if ev == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.pending[ev.RequestID]; ok {
		if e.entry.Response.StatusText == "" {
			e.entry.Response.StatusText = ev.ErrorText
		}
		e.entry.Time = durationMs(e.startMono, float64(ev.Timestamp))
		e.entry.Timings.Wait = e.entry.Time
		e.entry.Timings.Receive = 0
		r.complete(ev.RequestID, e)
	}
}

// complete moves an entry from pending to completed. Caller holds the lock.
func (r *Recorder) complete(id network.RequestID, e *entryState) {
	if e.finished {
		return
	}
	e.finished = true
	delete(r.pending, id)
	r.completed = append(r.completed, e)
}

// Build assembles the HAR. Still-pending requests (session ended mid-flight)
// are included so nothing is silently dropped. pageTitle is optional.
func (r *Recorder) Build(pageTitle string) *HAR {
	r.mu.Lock()
	defer r.mu.Unlock()

	all := make([]*entryState, 0, len(r.completed)+len(r.pending))
	all = append(all, r.completed...)
	for _, e := range r.pending {
		all = append(all, e)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].seq < all[j].seq })

	entries := make([]Entry, 0, len(all))
	pageStart := ""
	for _, e := range all {
		entry := e.entry
		entry.Pageref = "page_1"
		entries = append(entries, entry)
		if pageStart == "" {
			pageStart = entry.StartedDateTime
		}
	}
	if pageStart == "" {
		pageStart = epochToRFC3339(float64(time.Now().UnixNano()) / 1e9)
	}

	return &HAR{
		Log: Log{
			Version: harVersion,
			Creator: Creator{Name: creatorName, Version: harVersion},
			Pages: []Page{{
				StartedDateTime: pageStart,
				ID:              "page_1",
				Title:           pageTitle,
				PageTimings:     PageTimings{OnContentLoad: -1, OnLoad: -1},
			}},
			Entries: entries,
		},
	}
}

// WriteFile builds the HAR and writes it as indented JSON to path.
func (r *Recorder) WriteFile(path, pageTitle string) error {
	data, err := json.MarshalIndent(r.Build(pageTitle), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// EntryCount reports how many requests have been observed (pending + done).
func (r *Recorder) EntryCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.completed) + len(r.pending)
}

// --- helpers -------------------------------------------------------------

func applyResponse(e *Entry, resp *network.Response) {
	if resp == nil {
		return
	}
	e.Response.Status = resp.Status
	e.Response.StatusText = resp.StatusText
	e.Response.Headers = decodeHeaders(resp.Headers)
	e.Response.Content.MimeType = resp.MimeType
	if resp.EncodedDataLength > 0 {
		e.Response.BodySize = int64(resp.EncodedDataLength)
	}
	if resp.Protocol != nil && *resp.Protocol != "" {
		e.Response.HTTPVersion = *resp.Protocol
	}
	if resp.RemoteIPAddress != nil {
		e.ServerIPAddress = *resp.RemoteIPAddress
	}
}

func redirectLocation(resp *network.Response) string {
	for _, h := range decodeHeaders(resp.Headers) {
		if h.Name == "Location" || h.Name == "location" {
			return h.Value
		}
	}
	return ""
}

// decodeHeaders converts CDP Headers (raw JSON object) into HAR name/value pairs.
func decodeHeaders(h network.Headers) []NameValue {
	out := []NameValue{}
	if len(h) == 0 {
		return out
	}
	var m map[string]string
	if err := json.Unmarshal(h, &m); err != nil {
		return out
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, NameValue{Name: k, Value: m[k]})
	}
	return out
}

func queryString(rawURL string) []NameValue {
	out := []NameValue{}
	u, err := url.Parse(rawURL)
	if err != nil {
		return out
	}
	keys := make([]string, 0)
	q := u.Query()
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range q[k] {
			out = append(out, NameValue{Name: k, Value: v})
		}
	}
	return out
}

func postData(req *network.Request) *PostData {
	if req.PostData == nil || *req.PostData == "" {
		return nil
	}
	mime := ""
	if h := decodeHeaders(req.Headers); h != nil {
		for _, nv := range h {
			if nv.Name == "Content-Type" || nv.Name == "content-type" {
				mime = nv.Value
				break
			}
		}
	}
	return &PostData{MimeType: mime, Text: *req.PostData}
}

func bodySize(postData *string) int64 {
	if postData == nil {
		return 0
	}
	return int64(len(*postData))
}

// durationMs returns (end-start) in milliseconds, clamped to >= 0.
func durationMs(startMono, endMono float64) float64 {
	d := (endMono - startMono) * 1000
	if d < 0 {
		return 0
	}
	return d
}

// epochToRFC3339 converts seconds-since-epoch (CDP wall time) to RFC3339 millis.
func epochToRFC3339(epochSeconds float64) string {
	if epochSeconds <= 0 {
		return time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	}
	sec := int64(epochSeconds)
	nsec := int64((epochSeconds - float64(sec)) * 1e9)
	return time.Unix(sec, nsec).UTC().Format("2006-01-02T15:04:05.000Z07:00")
}
