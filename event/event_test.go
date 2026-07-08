package event

import (
	"sync"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

type initSpy struct {
	inited bool
}

func (s *initSpy) Init() {
	s.inited = true
}

type fileSpy struct {
	mu    sync.Mutex
	names []string
}

func (s *fileSpy) OnFileCreated(createdFile CreatedFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.names = append(s.names, createdFile.Name)
}

func TestInitIfNeeded(t *testing.T) {
	spy := &initSpy{}
	InitIfNeeded(spy)
	assert.True(t, spy.inited)
}

func TestFileCreatedNotifiesListener(t *testing.T) {
	spy := &fileSpy{}
	AddFileCreatedListener(spy)
	FileCreated(CreatedFile{Event: Event{RequestId: 1, SessionId: "s1"}, Name: "clip.png", Type: "image/png"})
	assert.Eventually(t, func() bool {
		spy.mu.Lock()
		defer spy.mu.Unlock()
		return len(spy.names) == 1 && spy.names[0] == "clip.png"
	}, time.Second, 10*time.Millisecond)
}
