// Package warm probes the selenoid-warm-pool orchestrator and exposes
// ready/total slot counts for hub /status (UI header WARM metric).
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

// Tracker periodically polls the warm-pool orchestrator.
type Tracker struct {
	baseURL string
	client  *http.Client
	mu      sync.RWMutex
	ready   int
	total   int
}

// NewTracker returns a tracker for the orchestrator base URL
// (e.g. http://127.0.0.1:9090). Empty URL is invalid — caller should skip.
func NewTracker(baseURL string) *Tracker {
	return &Tracker{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

// Snapshot returns the last known ready/total warm slot counts.
func (t *Tracker) Snapshot() (ready, total int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ready, t.total
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

type slotDTO struct {
	ReservedBy *string `json:"reservedBy"`
}

func (t *Tracker) refresh() {
	req, err := http.NewRequest(http.MethodGet, t.baseURL+"/pool/slots", nil)
	if err != nil {
		log.Printf("[-] [WARM_POOL] [request: %v]", err)
		return
	}
	resp, err := t.client.Do(req)
	if err != nil {
		t.mu.Lock()
		t.ready, t.total = 0, 0
		t.mu.Unlock()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.mu.Lock()
		t.ready, t.total = 0, 0
		t.mu.Unlock()
		return
	}
	var slots []slotDTO
	if err := json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		log.Printf("[-] [WARM_POOL] [decode: %v]", err)
		return
	}
	ready := 0
	for _, s := range slots {
		if s.ReservedBy == nil {
			ready++
		}
	}
	t.mu.Lock()
	t.ready = ready
	t.total = len(slots)
	t.mu.Unlock()
}
