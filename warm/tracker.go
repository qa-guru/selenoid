// Package warm probes the selenoid-warm-pool orchestrator and exposes
// ready/total slot counts for hub /status (UI header WARM / HOT metrics).
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
	baseURL   string
	client    *http.Client
	mu        sync.RWMutex
	warmReady int
	warmTotal int
	hotReady  int
	hotTotal  int
}

// NewTracker returns a tracker for the orchestrator base URL
// (e.g. http://127.0.0.1:9090). Empty URL is invalid — caller should skip.
func NewTracker(baseURL string) *Tracker {
	return &Tracker{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 3 * time.Second},
	}
}

// Snapshot returns the last known warm and hot ready/total slot counts.
func (t *Tracker) Snapshot() (warmReady, warmTotal, hotReady, hotTotal int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.warmReady, t.warmTotal, t.hotReady, t.hotTotal
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
	Pool       string  `json:"pool"`
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
	var slots []slotDTO
	if err := json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		log.Printf("[-] [WARM_POOL] [decode: %v]", err)
		return
	}
	warmReady, warmTotal, hotReady, hotTotal := 0, 0, 0, 0
	for _, s := range slots {
		if strings.EqualFold(s.Pool, "hot") {
			hotTotal++
			if s.ReservedBy == nil {
				hotReady++
			}
			continue
		}
		warmTotal++
		if s.ReservedBy == nil {
			warmReady++
		}
	}
	t.mu.Lock()
	t.warmReady, t.warmTotal = warmReady, warmTotal
	t.hotReady, t.hotTotal = hotReady, hotTotal
	t.mu.Unlock()
}

func (t *Tracker) clear() {
	t.mu.Lock()
	t.warmReady, t.warmTotal = 0, 0
	t.hotReady, t.hotTotal = 0, 0
	t.mu.Unlock()
}
