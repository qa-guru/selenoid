package upload

import (
	"testing"
	"time"

	"github.com/aerokube/selenoid/event"
	assert "github.com/stretchr/testify/require"
)

type mockUploader struct {
	done chan string
}

func (m *mockUploader) Upload(createdFile event.CreatedFile) (bool, error) {
	m.done <- createdFile.Name
	return true, nil
}

func TestOnFileCreatedInvokesUploader(t *testing.T) {
	mock := &mockUploader{done: make(chan string, 1)}
	ul := &Upload{uploaders: []Uploader{mock}}
	ul.OnFileCreated(event.CreatedFile{
		Event: event.Event{RequestId: 42, SessionId: "sid"},
		Name:  "test.log",
		Type:  "text/plain",
	})
	select {
	case name := <-mock.done:
		assert.Equal(t, "test.log", name)
	case <-time.After(time.Second):
		t.Fatal("uploader was not invoked within 1s")
	}
}
