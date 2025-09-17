package sync

// Semaphore is a weighted semaphore implementation built around [Chan].
type Semaphore struct {
	ch Chan[empty]
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

// Acquire acquires a semaphore of weight n. If the size given was zero, this
// operation is a nop.
//
// The weight provided cannot be greater than the semaphore capacity. If it is,
// the semaphore capacity is used instead.
func (s *Semaphore) Acquire(n int) {
	s.ch.Send(make([]empty, min(s.ch.Cap(), n))...)
}

// TryAcquire attempts to acquire a semaphore of weight n. It returns
// immediately if the semaphore cannot be acquired. If the size given was zero,
// this operation is a nop.
//
// The weight provided cannot be greater than the semaphore capacity. If it is,
// the semaphore capacity is used instead.
func (s *Semaphore) TryAcquire(n int) bool {
	for range min(s.ch.Cap(), n) {
		select {
		case s.ch.Load() <- empty{}:
		default:
			return false
		}
	}
	return true
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
