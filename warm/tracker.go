// Package warm probes the selenoid-pool orchestrator and exposes
// ready/total slot counts plus slot rows for hub /status (UI header + Statistics).
package warm

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Slot is a pool row on hub /status (no URLs).
type Slot struct {
	ID         string  `json:"id"`
	Browser    string  `json:"browser"`
	Protocol   string  `json:"protocol"`
	Pool       string  `json:"pool"`
	ReservedBy *string `json:"reservedBy"`
}

// Snapshot is the last known warm/hot slot counts and rows.
type Snapshot struct {
	WarmReady int
	WarmTotal int
	HotReady  int
	HotTotal  int
	WarmSlots []Slot
	HotSlots  []Slot
}

// Tracker periodically polls the warm-pool orchestrator.
type Tracker struct {
	baseURL   string
	client    *http.Client
	mu        sync.RWMutex
	warmReady int
	warmTotal int
	hotReady  int
	hotTotal  int
	warmSlots []Slot
	hotSlots  []Slot
}

// NewTracker returns a tracker for the orchestrator base URL
// (e.g. http://127.0.0.1:9090). Empty URL is invalid — caller should skip.
func NewTracker(baseURL string) *Tracker {
	return &Tracker{
		baseURL:   strings.TrimRight(baseURL, "/"),
		client:    &http.Client{Timeout: 3 * time.Second},
		warmSlots: []Slot{},
		hotSlots:  []Slot{},
	}
}

func copySlots(in []Slot) []Slot {
	out := make([]Slot, len(in))
	copy(out, in)
	return out
}

// Snapshot returns the last known warm and hot ready/total counts and slot rows.
func (t *Tracker) Snapshot() Snapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return Snapshot{
		WarmReady: t.warmReady,
		WarmTotal: t.warmTotal,
		HotReady:  t.hotReady,
		HotTotal:  t.hotTotal,
		WarmSlots: copySlots(t.warmSlots),
		HotSlots:  copySlots(t.hotSlots),
	}
}

// Start polls until ctx is cancelled. Safe to call once in a goroutine.
func (t *Tracker) Start(ctx context.Context) {
	t.refresh()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.refresh()
		}
	}
}

func (t *Tracker) refresh() {
	req, err := http.NewRequest(http.MethodGet, t.baseURL+"/pool/slots", nil)
	if err != nil {
		log.Printf("[-] [WARM_POOL] [request: %v]", err)
		return
	}
	resp, err := t.client.Do(req)
	if err != nil {
		t.clear()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.clear()
		return
	}
	var slots []Slot
	if err := json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		log.Printf("[-] [WARM_POOL] [decode: %v]", err)
		return
	}
	warmReady, warmTotal, hotReady, hotTotal := 0, 0, 0, 0
	warmSlots, hotSlots := make([]Slot, 0), make([]Slot, 0)
	for _, s := range slots {
		if strings.EqualFold(s.Pool, "hot") {
			hotTotal++
			if s.ReservedBy == nil {
				hotReady++
			}
			hotSlots = append(hotSlots, s)
			continue
		}
		warmTotal++
		if s.ReservedBy == nil {
			warmReady++
		}
		warmSlots = append(warmSlots, s)
	}
	t.mu.Lock()
	t.warmReady, t.warmTotal = warmReady, warmTotal
	t.hotReady, t.hotTotal = hotReady, hotTotal
	t.warmSlots, t.hotSlots = warmSlots, hotSlots
	t.mu.Unlock()
}

func (t *Tracker) clear() {
	t.mu.Lock()
	t.warmReady, t.warmTotal = 0, 0
	t.hotReady, t.hotTotal = 0, 0
	t.warmSlots, t.hotSlots = []Slot{}, []Slot{}
	t.mu.Unlock()
}
