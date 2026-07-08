package jsonerror

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestSeleniumErrorEncode(t *testing.T) {
	rec := httptest.NewRecorder()
	err := errors.New("boom")
	InvalidSessionID(err).Encode(rec)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	value := body["value"].(map[string]interface{})
	assert.Equal(t, "invalid session id", value["error"])
	assert.Equal(t, "boom", value["message"])
}

func TestSeleniumErrorString(t *testing.T) {
	se := SessionNotCreated(errors.New("docker down"))
	assert.Contains(t, se.Error(), "session not created")
	assert.Contains(t, se.Error(), "docker down")
}
