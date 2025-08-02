package backoff

import (
	"time"

	"go.chrisrx.dev/x/sync"
)

// A Ticker is like [time.Ticker] but accepts a [Backoff].
type Ticker struct {
	time.Ticker
	c sync.Chan[time.Time]
	b Backoff
}

// NewTicker constructs a new [Ticker] with the provided [Backoff].
func NewTicker(b Backoff) *Ticker {
	return &Ticker{b: b}
}

// Stop stops the ticker by closing the underlying ticker channel. It is safe
// to call multiple times.
func (t *Ticker) Stop() {
	t.c.Close()
	t.b.Reset()
}

// Next returns a receive-only channel that produces at intervals determined by
// the configured [Backoff].
func (t *Ticker) Next() <-chan time.Time {
	ch, isNew := t.c.LoadOrNew()
	if isNew {
		go func() {
			for {
				select {
				case <-t.c.Recv():
					return
				case <-time.After(t.b.Next()):
					t.c.Send(time.Now())
				}
			}
		}()
	}
	return ch
}
