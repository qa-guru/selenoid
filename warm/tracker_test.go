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
	ready, total := tr.Snapshot()
	if total != 2 {
		t.Fatalf("total=%d want 2", total)
	}
	if ready != 1 {
		t.Fatalf("ready=%d want 1", ready)
	}
}

func TestTrackerDownClearsCounts(t *testing.T) {
	tr := NewTracker("http://127.0.0.1:1")
	tr.mu.Lock()
	tr.ready, tr.total = 2, 2
	tr.mu.Unlock()
	tr.refresh()
	ready, total := tr.Snapshot()
	if ready != 0 || total != 0 {
		t.Fatalf("got ready=%d total=%d want 0/0", ready, total)
	}
}
