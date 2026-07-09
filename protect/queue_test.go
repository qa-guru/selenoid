package protect

import (
	"context"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestQueueCreateReleaseCounters(t *testing.T) {
	q := New(2, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.True(t, q.Wait(ctx))

	q.Create()
	assert.Equal(t, 1, q.Used())
	assert.Equal(t, 0, q.Pending())

	q.Release()
	assert.Equal(t, 0, q.Used())
}

func TestQueueDrop(t *testing.T) {
	q := New(1, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.True(t, q.Wait(ctx))
	assert.Equal(t, 1, q.Pending())

	q.Drop()
	assert.Equal(t, 0, q.Pending())
	assert.Equal(t, 0, q.Used())
}

func TestQueueWaitCanceled(t *testing.T) {
	// Limit 0: limit send blocks, so canceled context wins (no select race with buffered limit).
	q := New(0, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, q.Wait(ctx))
}
