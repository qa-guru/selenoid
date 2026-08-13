package warm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientReserveAndRelease(t *testing.T) {
	var released string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/pool/reserve":
			var in map[string]any
			_ = json.Unmarshal(body, &in)
			if in["loopback"] != true || in["browser"] != "chrome" || in["protocol"] != "webdriver" {
				t.Errorf("reserve body=%s", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"slot":{"id":"pool-chrome-1","webdriverUrl":"http://127.0.0.1:14441/"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/pool/release":
			var in map[string]string
			_ = json.Unmarshal(body, &in)
			released = in["slotId"]
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"slotId":"pool-chrome-1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	id, wd, err := c.Reserve("webdriver", "chrome", "hub-1", true)
	if err != nil {
		t.Fatal(err)
	}
	if id != "pool-chrome-1" || wd != "http://127.0.0.1:14441/" {
		t.Fatalf("id=%q wd=%q", id, wd)
	}
	if err := c.Release(id); err != nil {
		t.Fatal(err)
	}
	if released != "pool-chrome-1" {
		t.Fatalf("released=%q", released)
	}
}

func TestClientReserve409(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"no available slots"}`))
	}))
	defer srv.Close()

	_, _, err := NewClient(srv.URL).Reserve("webdriver", "chrome", "hub-1", true)
	if err != ErrNoSlot {
		t.Fatalf("err=%v want ErrNoSlot", err)
	}
}

func TestClientReserveDown(t *testing.T) {
	_, _, err := NewClient("http://127.0.0.1:1").Reserve("webdriver", "chrome", "hub-1", true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsLoopbackURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"http://127.0.0.1:14441/", true},
		{"http://localhost:14441/wd/hub", true},
		{"http://[::1]:14441/", true},
		{"http://warm-chrome-1:4444/", false},
		{"http://0.0.0.0:4444/", false},
		{"", false},
		{"not a url", false},
	}
	for _, tc := range cases {
		if got := IsLoopbackURL(tc.in); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.in, got, tc.want)
		}
	}
}
