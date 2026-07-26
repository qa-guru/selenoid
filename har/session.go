package har

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/rpcc"
)

// dialTimeout bounds how long we wait for the DevTools websocket to accept us.
const dialTimeout = 5 * time.Second

// stopTimeout bounds how long Stop waits for the event goroutines to drain.
const stopTimeout = 3 * time.Second

// Session is a live HAR capture over a Chrome DevTools Protocol connection.
//
// It dials the browser container DevTools endpoint (the same one exposed via
// `/devtools/<session-id>/page` and the `se:cdp` capability, browser port
// 7070), enables the Network domain and streams events into a Recorder until
// Stop is called.
type Session struct {
	recorder *Recorder
	conn     *rpcc.Conn
	cancel   context.CancelFunc
	done     chan struct{}
}

// Start opens a CDP connection to wsURL (e.g. ws://<host:7070>/page), enables
// the Network domain and begins recording. The passed ctx bounds only the
// initial dial/enable; recording then runs until Stop.
func Start(ctx context.Context, wsURL string) (*Session, error) {
	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	defer dialCancel()

	conn, err := rpcc.DialContext(dialCtx, wsURL)
	if err != nil {
		return nil, fmt.Errorf("dial devtools %s: %w", wsURL, err)
	}

	c := cdp.NewClient(conn)
	runCtx, cancel := context.WithCancel(context.Background())
	if err := c.Network.Enable(runCtx, network.NewEnableArgs()); err != nil {
		cancel()
		_ = conn.Close()
		return nil, fmt.Errorf("enable network domain: %w", err)
	}

	s := &Session{
		recorder: NewRecorder(),
		conn:     conn,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
	go s.run(runCtx, c)
	return s, nil
}

// run subscribes to the four network events that make up a HAR entry and
// pumps them into the recorder until the context is canceled.
func (s *Session) run(ctx context.Context, c *cdp.Client) {
	defer close(s.done)

	var wg sync.WaitGroup
	spawn := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	if cl, err := c.Network.RequestWillBeSent(ctx); err == nil {
		spawn(func() {
			defer cl.Close()
			for {
				ev, err := cl.Recv()
				if err != nil {
					return
				}
				s.recorder.RequestWillBeSent(ev)
			}
		})
	}
	if cl, err := c.Network.ResponseReceived(ctx); err == nil {
		spawn(func() {
			defer cl.Close()
			for {
				ev, err := cl.Recv()
				if err != nil {
					return
				}
				s.recorder.ResponseReceived(ev)
			}
		})
	}
	if cl, err := c.Network.LoadingFinished(ctx); err == nil {
		spawn(func() {
			defer cl.Close()
			for {
				ev, err := cl.Recv()
				if err != nil {
					return
				}
				s.recorder.LoadingFinished(ev)
			}
		})
	}
	if cl, err := c.Network.LoadingFailed(ctx); err == nil {
		spawn(func() {
			defer cl.Close()
			for {
				ev, err := cl.Recv()
				if err != nil {
					return
				}
				s.recorder.LoadingFailed(ev)
			}
		})
	}

	wg.Wait()
}

// Stop ends recording, closes the CDP connection and returns the Recorder so
// the caller can serialize the collected HAR. It is safe to call once.
func (s *Session) Stop() *Recorder {
	s.cancel()
	select {
	case <-s.done:
	case <-time.After(stopTimeout):
	}
	_ = s.conn.Close()
	return s.recorder
}
