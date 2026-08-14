package warm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrackerSnapshotFromSlots(t *testing.T) {
	owner := "jenkins"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/slots" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "pool-chrome-1", "reservedBy": nil},
			{"id": "pool-chrome-2", "reservedBy": owner},
		})
	}))
	defer srv.Close()

	tr := NewTracker(srv.URL)
	tr.refresh()
	warmReady, warmTotal, hotReady, hotTotal := tr.Snapshot()
	if warmTotal != 2 {
		t.Fatalf("warmTotal=%d want 2", warmTotal)
	}
	if warmReady != 1 {
		t.Fatalf("warmReady=%d want 1", warmReady)
	}
	if hotReady != 0 || hotTotal != 0 {
		t.Fatalf("hot=%d/%d want 0/0", hotReady, hotTotal)
	}
}

func TestTrackerSplitsHotFromWarm(t *testing.T) {
	owner := "jenkins"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pool/slots" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "pool-chrome-1", "pool": "warm", "reservedBy": nil},
			{"id": "pool-chrome-2", "pool": "warm", "reservedBy": owner},
			{"id": "pool-pw-1", "pool": "warm", "reservedBy": nil},
			{"id": "pool-pw-min-1", "pool": "warm", "reservedBy": nil},
			{"id": "pool-hot-chrome-min-1", "pool": "hot", "reservedBy": nil},
			{"id": "pool-hot-pw-min-1", "pool": "hot", "reservedBy": owner},
		})
	}))
	defer srv.Close()

	tr := NewTracker(srv.URL)
	tr.refresh()
	warmReady, warmTotal, hotReady, hotTotal := tr.Snapshot()
	if warmReady != 3 || warmTotal != 4 {
		t.Fatalf("warm=%d/%d want 3/4", warmReady, warmTotal)
	}
	if hotReady != 1 || hotTotal != 2 {
		t.Fatalf("hot=%d/%d want 1/2", hotReady, hotTotal)
	}
	if warmTotal == hotTotal {
		t.Fatal("hot count must not collapse into warm")
	}
}

func TestTrackerDownClearsCounts(t *testing.T) {
	tr := NewTracker("http://127.0.0.1:1")
	tr.mu.Lock()
	tr.warmReady, tr.warmTotal = 2, 2
	tr.hotReady, tr.hotTotal = 1, 2
	tr.mu.Unlock()
	tr.refresh()
	warmReady, warmTotal, hotReady, hotTotal := tr.Snapshot()
	if warmReady != 0 || warmTotal != 0 || hotReady != 0 || hotTotal != 0 {
		t.Fatalf("got warm=%d/%d hot=%d/%d want 0/0", warmReady, warmTotal, hotReady, hotTotal)
	}
}
