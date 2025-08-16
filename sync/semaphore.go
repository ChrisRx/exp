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
	// The requested weight is higher than the maximum capacity of this
	// semaphore. This is a mistake and will always cause a deadlock, so the
	// weight is set to the maximum capacity.
	if s.ch.Cap() < n {
		n = s.ch.Cap()
	}
	s.ch.Send(make([]struct{}, n)...)
}

// Release releases a semaphore. If the size given was zero this operation is a
// nop.
func (s *Semaphore) Release() {
	if s.ch.Cap() == 0 {
		return
	}
	s.release()
}

func (s *Semaphore) release() {
	select {
	case <-s.ch.Recv():
	default:
	}
}

// ReleaseN releases a semaphore with the provided weight. If the size given
// was zero this operation is a nop.
func (s *Semaphore) ReleaseN(n int) {
	if s.ch.Cap() == 0 {
		return
	}
	for range n {
		s.release()
	}
}
