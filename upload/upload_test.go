package upload

import (
	"testing"
	"time"

	"github.com/aerokube/selenoid/event"
	assert "github.com/stretchr/testify/require"
)

type mockUploader struct {
	uploaded []string
}

func (m *mockUploader) Upload(createdFile event.CreatedFile) (bool, error) {
	m.uploaded = append(m.uploaded, createdFile.Name)
	return true, nil
}

func TestOnFileCreatedInvokesUploader(t *testing.T) {
	mock := &mockUploader{}
	ul := &Upload{uploaders: []Uploader{mock}}
	ul.OnFileCreated(event.CreatedFile{
		Event: event.Event{RequestId: 42, SessionId: "sid"},
		Name:  "test.log",
		Type:  "text/plain",
	})
	assert.Eventually(t, func() bool {
		return len(mock.uploaded) == 1 && mock.uploaded[0] == "test.log"
	}, time.Second, 10*time.Millisecond)
}
