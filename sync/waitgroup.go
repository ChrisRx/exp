package sync

import (
	"sync"
)

// A BoundedWaitGroup works like a [sync.WaitGroup] with a semaphore that
// limits concurrency.
type BoundedWaitGroup struct {
	limit *Semaphore
	wg    sync.WaitGroup
}

// Add adds delta, which may be negative, to the [WaitGroup] counter.
func (w *BoundedWaitGroup) Add(delta int) {
	if w.limit != nil {
		w.limit.Acquire(delta)
	}
	w.wg.Add(delta)
}

// Done decrements the [WaitGroup] counter by one.
func (w *BoundedWaitGroup) Done() {
	if w.limit != nil {
		w.limit.Release()
	}
	w.wg.Done()
}

// SetLimit sets the bounds for concurrency to the provided value.
func (w *BoundedWaitGroup) SetLimit(n int) {
	w.limit = NewSemaphore(n)
}

// Wait blocks until the [WaitGroup] counter is zero.
func (w *BoundedWaitGroup) Wait() {
	w.wg.Wait()
}
