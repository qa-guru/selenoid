package warm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNoSlot is returned when the orchestrator has no matching free slot (HTTP 409).
var ErrNoSlot = errors.New("no available warm slots")

// Client calls the selenoid-pool orchestrator (reserve / release).
type Client struct {
	baseURL  string
	http     *http.Client
	leaseSec int
}

// NewClient talks to the orchestrator base URL (e.g. http://127.0.0.1:9090).
// leaseSec is the reservation TTL sent as ttlSec on reserve — it must outlive
// the longest possible hub session so the pool reaper cannot free a slot that
// is still in use (the orchestrator clamps it to its max_lease_ttl).
func NewClient(baseURL string, leaseSec int) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		http:     &http.Client{Timeout: 2 * time.Second},
		leaseSec: leaseSec,
	}
}

// Reserve claims a slot. loopback=true asks for a host-reachable WebDriver URL.
func (c *Client) Reserve(protocol, browser, owner string, loopback bool) (slotID, webdriverURL string, err error) {
	if c == nil || c.baseURL == "" {
		return "", "", ErrNoSlot
	}
	body := map[string]any{
		"protocol": protocol,
		"browser":  browser,
		"owner":    owner,
		"loopback": loopback,
	}
	if c.leaseSec > 0 {
		body["ttlSec"] = c.leaseSec
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/pool/reserve", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode == http.StatusConflict {
		return "", "", ErrNoSlot
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("reserve HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		OK   bool `json:"ok"`
		Slot struct {
			ID           string `json:"id"`
			WebdriverURL string `json:"webdriverUrl"`
		} `json:"slot"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", "", err
	}
	return out.Slot.ID, out.Slot.WebdriverURL, nil
}

// Release frees a slot (best-effort reset on the orchestrator). The owner must
// match the reserve-time owner — the orchestrator rejects foreign releases.
func (c *Client) Release(slotID, owner string) error {
	if c == nil || c.baseURL == "" || slotID == "" {
		return nil
	}
	payload, err := json.Marshal(map[string]string{"slotId": slotID, "owner": owner})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/pool/release", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("release HTTP %d", resp.StatusCode)
	}
	return nil
}

// IsLoopbackURL reports whether raw is a host-loopback http(s) URL the hub binary can dial.
func IsLoopbackURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}
