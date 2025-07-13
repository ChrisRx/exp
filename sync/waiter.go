package sync

// A Waiter uses a [LazyChan] to wait for an event to occur.
type Waiter struct {
	ch   LazyChan[struct{}]
	done OnceAgain
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
	w.done.Reset()
}

// Wait blocks until receiving a done signal.
func (w *Waiter) Wait() {
	<-w.ch.Load()
}
