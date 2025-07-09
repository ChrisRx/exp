package sync

import (
	"sync"
)

// A Waiter uses a [Chan] to wait for an event to occur.
type Waiter struct {
	ch   Chan[struct{}]
	done sync.Once
}

// Done sends a done signal for any waiters currently blocked.
func (w *Waiter) Done() {
	w.done.Do(func() {
		w.ch.Close()
	})
}

// Reset resets the waiter to allow reuse. Any waiters currently blocked will
// release immediately.
func (w *Waiter) Reset() {
	w.ch.Reset()
	w.done = sync.Once{}
}

// Wait blocks until receiving a done signal.
func (w *Waiter) Wait() {
	<-w.ch.Load()
}
