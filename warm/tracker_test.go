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
			{"id": "pool-chrome-1", "browser": "chrome", "protocol": "webdriver", "reservedBy": nil},
			{"id": "pool-chrome-2", "browser": "chrome", "protocol": "webdriver", "reservedBy": owner},
		})
	}))
	defer srv.Close()

	tr := NewTracker(srv.URL)
	tr.refresh()
	snap := tr.Snapshot()
	if snap.WarmTotal != 2 {
		t.Fatalf("warmTotal=%d want 2", snap.WarmTotal)
	}
	if snap.WarmReady != 1 {
		t.Fatalf("warmReady=%d want 1", snap.WarmReady)
	}
	if snap.HotReady != 0 || snap.HotTotal != 0 {
		t.Fatalf("hot=%d/%d want 0/0", snap.HotReady, snap.HotTotal)
	}
	if len(snap.WarmSlots) != 2 || len(snap.HotSlots) != 0 {
		t.Fatalf("slots warm=%d hot=%d", len(snap.WarmSlots), len(snap.HotSlots))
	}
	if snap.WarmSlots[0].Browser != "chrome" || snap.WarmSlots[1].ReservedBy == nil {
		t.Fatalf("warm rows=%+v", snap.WarmSlots)
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
			{"id": "pool-chrome-1", "pool": "warm", "browser": "chrome", "protocol": "webdriver", "reservedBy": nil},
			{"id": "pool-chrome-2", "pool": "warm", "browser": "chrome", "protocol": "webdriver", "reservedBy": owner},
			{"id": "pool-pw-1", "pool": "warm", "browser": "chromium", "protocol": "playwright", "reservedBy": nil},
			{"id": "pool-pw-min-1", "pool": "warm", "browser": "chromium", "protocol": "playwright", "reservedBy": nil},
			{"id": "pool-hot-chrome-min-1", "pool": "hot", "browser": "chrome", "protocol": "webdriver", "reservedBy": nil},
			{"id": "pool-hot-pw-min-1", "pool": "hot", "browser": "chromium", "protocol": "playwright", "reservedBy": owner},
		})
	}))
	defer srv.Close()

	tr := NewTracker(srv.URL)
	tr.refresh()
	snap := tr.Snapshot()
	if snap.WarmReady != 3 || snap.WarmTotal != 4 {
		t.Fatalf("warm=%d/%d want 3/4", snap.WarmReady, snap.WarmTotal)
	}
	if snap.HotReady != 1 || snap.HotTotal != 2 {
		t.Fatalf("hot=%d/%d want 1/2", snap.HotReady, snap.HotTotal)
	}
	if snap.WarmTotal == snap.HotTotal {
		t.Fatal("hot count must not collapse into warm")
	}
	if len(snap.WarmSlots) != 4 || len(snap.HotSlots) != 2 {
		t.Fatalf("slot rows warm=%d hot=%d", len(snap.WarmSlots), len(snap.HotSlots))
	}
	if snap.HotSlots[0].ID != "pool-hot-chrome-min-1" || snap.HotSlots[1].Pool != "hot" {
		t.Fatalf("hot rows=%+v", snap.HotSlots)
	}
}

func TestTrackerDownClearsCounts(t *testing.T) {
	tr := NewTracker("http://127.0.0.1:1")
	owner := "x"
	tr.mu.Lock()
	tr.warmReady, tr.warmTotal = 2, 2
	tr.hotReady, tr.hotTotal = 1, 2
	tr.warmSlots = []Slot{{ID: "w", Browser: "chrome"}}
	tr.hotSlots = []Slot{{ID: "h", Browser: "chrome", ReservedBy: &owner}}
	tr.mu.Unlock()
	tr.refresh()
	snap := tr.Snapshot()
	if snap.WarmReady != 0 || snap.WarmTotal != 0 || snap.HotReady != 0 || snap.HotTotal != 0 {
		t.Fatalf("got warm=%d/%d hot=%d/%d want 0/0", snap.WarmReady, snap.WarmTotal, snap.HotReady, snap.HotTotal)
	}
	if len(snap.WarmSlots) != 0 || len(snap.HotSlots) != 0 {
		t.Fatalf("cleared slots warm=%d hot=%d", len(snap.WarmSlots), len(snap.HotSlots))
	}
}
