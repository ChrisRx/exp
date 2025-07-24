package sync

import (
	"sync"
)

// A WaitGroup works like a [sync.WaitGroup] with a semaphore that limits
// concurrency.
type WaitGroup struct {
	limit Semaphore
	wg    sync.WaitGroup
}

// Add adds delta, which may be negative, to the [WaitGroup] counter.
func (w *WaitGroup) Add(delta int) {
	w.limit.Acquire(delta)
	w.wg.Add(delta)
}

// Done decrements the [WaitGroup] counter by one.
func (w *WaitGroup) Done() {
	w.limit.Release()
	w.wg.Done()
}

// SetLimit sets the bounds for concurrency to the provided value.
func (w *WaitGroup) SetLimit(n int) {
	w.limit.SetLimit(n)
}

// Wait blocks until the [WaitGroup] counter is zero.
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}
