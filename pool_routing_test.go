package main

import (
	"encoding/json"
	"testing"

	"github.com/qa-guru/selenoid/session"
)

func TestRequestedPoolResolution(t *testing.T) {
	cases := []struct {
		name string
		caps session.Caps
		want string
	}{
		{"unset", session.Caps{}, ""},
		{"options warm", session.Caps{Pool: "warm"}, "warm"},
		{"options case", session.Caps{Pool: " WARM "}, "warm"},
		{"labels cold", session.Caps{Labels: map[string]string{"pool": "cold"}}, "cold"},
		{"options beats labels", session.Caps{Pool: "warm", Labels: map[string]string{"pool": "cold"}}, "warm"},
		{"raw invalid", session.Caps{Pool: "lukewarm"}, "lukewarm"},
	}
	for _, tc := range cases {
		if got := tc.caps.RequestedPool(); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestValidatePoolSelector(t *testing.T) {
	for _, ok := range []string{"", "auto", "cold", "warm"} {
		c := &session.Caps{Pool: ok}
		if err := validatePoolSelector(c); err != nil {
			t.Errorf("pool=%q must be valid: %v", ok, err)
		}
	}
	hot := &session.Caps{Pool: "hot"}
	if err := validatePoolSelector(hot); err == nil || !contains(err.Error(), "pool/lease") {
		t.Errorf("hot must refuse with lease hint, got %v", err)
	}
	bad := &session.Caps{Pool: "lukewarm"}
	if err := validatePoolSelector(bad); err == nil || !contains(err.Error(), "lukewarm") {
		t.Errorf("invalid value must name itself, got %v", err)
	}
	// labels.pool goes through the same validation.
	viaLabels := &session.Caps{Labels: map[string]string{"pool": "hot"}}
	if err := validatePoolSelector(viaLabels); err == nil {
		t.Error("labels.pool=hot must refuse too")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestProcessBodyInjectsPoolMeta(t *testing.T) {
	in := `{"value":{"sessionId":"abc123","capabilities":{"browserName":"chrome","browserVersion":"152.0.1"}}}`
	out, sid, err := processBody([]byte(in), "selenoid.example", map[string]interface{}{
		"servedPool":    "warm",
		"requestedPool": "auto",
		"servedSlot":    "pool-chrome-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sid != "abc123" {
		t.Fatalf("sid=%q", sid)
	}
	var body map[string]any
	if err := json.Unmarshal(out, &body); err != nil {
		t.Fatal(err)
	}
	caps := body["value"].(map[string]any)["capabilities"].(map[string]any)
	if caps["servedPool"] != "warm" || caps["requestedPool"] != "auto" || caps["servedSlot"] != "pool-chrome-1" {
		t.Fatalf("caps=%v", caps)
	}
	if caps["se:cdp"] != "ws://selenoid.example/devtools/abc123/" {
		t.Fatalf("se:cdp=%v", caps["se:cdp"])
	}
	if caps["se:cdpVersion"] != "152.0.1" {
		t.Fatalf("se:cdpVersion=%v", caps["se:cdpVersion"])
	}
}

func TestRemoveSelenoidOptionsStripsPool(t *testing.T) {
	in := `{"capabilities":{"alwaysMatch":{"browserName":"chrome","pool":"warm","selenoid:options":{"enableVNC":true}},"firstMatch":[{"pool":"cold","selenoid:options":{}}]},"desiredCapabilities":{"pool":"warm","selenoid:options":{}}}`
	out := removeSelenoidOptions([]byte(in))
	var body map[string]any
	if err := json.Unmarshal(out, &body); err != nil {
		t.Fatal(err)
	}
	am := body["capabilities"].(map[string]any)["alwaysMatch"].(map[string]any)
	if _, ok := am["pool"]; ok {
		t.Fatal("alwaysMatch still has pool")
	}
	if _, ok := am["selenoid:options"]; ok {
		t.Fatal("alwaysMatch still has selenoid:options")
	}
	fm := body["capabilities"].(map[string]any)["firstMatch"].([]any)[0].(map[string]any)
	if _, ok := fm["pool"]; ok {
		t.Fatal("firstMatch still has pool")
	}
	dc := body["desiredCapabilities"].(map[string]any)
	if _, ok := dc["pool"]; ok {
		t.Fatal("desiredCapabilities still has pool")
	}
}

func TestPoolRoutingStatsRecord(t *testing.T) {
	s := &poolRoutingStats{}
	s.record("s1", "warm", "warm", "pool-chrome-1", "")
	s.record("s2", "", "cold", "", "warm-unavailable")
	s.record("s3", "cold", "cold", "", "")
	s.record("s4", "auto", "cold", "", "sidecar-caps")

	snap := s.snapshot()
	if snap.WarmServed != 1 || snap.ColdServed != 3 {
		t.Fatalf("counters warm=%d cold=%d", snap.WarmServed, snap.ColdServed)
	}
	// Only the runtime miss counts: sidecar-caps is expected cold, not a fallback.
	if snap.WarmFallbacks != 1 {
		t.Fatalf("fallbacks=%d", snap.WarmFallbacks)
	}
	if len(snap.Recent) != 4 {
		t.Fatalf("recent=%d", len(snap.Recent))
	}
	// Newest first.
	if snap.Recent[0].Session != "s4" || snap.Recent[0].Reason != "sidecar-caps" {
		t.Fatalf("recent[0]=%+v", snap.Recent[0])
	}
	if snap.Recent[1].Requested != "cold" || snap.Recent[1].Reason != "" {
		t.Fatalf("recent[1]=%+v", snap.Recent[1])
	}
	if snap.Recent[2].Requested != "auto" {
		t.Fatalf("unset requested must normalize to auto: %+v", snap.Recent[2])
	}
	if snap.Recent[3].Requested != "warm" || snap.Recent[3].Slot != "pool-chrome-1" {
		t.Fatalf("recent[3]=%+v", snap.Recent[3])
	}
}

func TestPoolRoutingStatsCapsAtKeep(t *testing.T) {
	s := &poolRoutingStats{}
	for i := 0; i < poolRoutesKeep+5; i++ {
		s.record("s", "cold", "cold", "", "")
	}
	snap := s.snapshot()
	if len(snap.Recent) != poolRoutesKeep {
		t.Fatalf("recent=%d want %d", len(snap.Recent), poolRoutesKeep)
	}
}
