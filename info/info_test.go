package info

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestRequestInfoBasicAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/status", nil)
	req.SetBasicAuth("alice", "secret")
	user, remote := RequestInfo(req)
	assert.Equal(t, "alice", user)
	assert.NotEmpty(t, remote)
}

func TestRequestInfoForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/status", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	user, remote := RequestInfo(req)
	assert.Equal(t, "unknown", user)
	assert.Equal(t, "203.0.113.1", remote)
}

func TestSecondsSince(t *testing.T) {
	start := time.Now().Add(-2 * time.Second)
	elapsed := SecondsSince(start)
	assert.GreaterOrEqual(t, elapsed, 1.9)
}
