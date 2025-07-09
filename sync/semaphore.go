package sync

// Semaphore is a weighted semaphore implementation built around [Chan].
type Semaphore struct {
	ch Chan[struct{}]
}

// NewSemaphore constructs a new semaphore using the provided size.
func NewSemaphore(n int) *Semaphore {
	s := &Semaphore{}
	s.SetLimit(n)
	return s
}

// SetLimit sets the limit for the semaphore.
func (s *Semaphore) SetLimit(n int) {
	s.ch.New(n)
}

// Acquire acquires a semaphore of weight n. If the size given was zero this
// operation is a nop.
func (s *Semaphore) Acquire(n int) {
	if s.ch.Cap() == 0 {
		return
	}
	for range n {
		s.ch.Load() <- struct{}{}
	}
}

// Release releases a semaphore. If the size given was zero this operation is a
// nop.
func (s *Semaphore) Release() {
	if s.ch.Cap() == 0 {
		return
	}
	<-s.ch.Load()
}
