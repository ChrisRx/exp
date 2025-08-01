package backoff

import (
	"time"

	"go.chrisrx.dev/x/sync"
)

type Ticker struct {
	c sync.Chan[time.Time]
	b Backoff
}

func NewTicker(b Backoff) *Ticker {
	return &Ticker{b: b}
}

func (t *Ticker) Stop() {
	t.c.Close()
	t.b.Reset()
}

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
