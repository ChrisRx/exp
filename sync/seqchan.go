package sync

import (
	"iter"
	"time"

	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/safe"
)

// SeqChan wraps a [Chan], providing Send/Recv methods that operate on iterator
// sequences.
type SeqChan[T any] struct {
	Chan[T]
}

// NewSeqChan constructs a new [SeqChan] of type T with the provided capacity.
func NewSeqChan[T any](capacity int) *SeqChan[T] {
	return &SeqChan[T]{
		Chan: ptr.From(NewChan[T](capacity)),
	}
}

// Send works like [Chan.Send] but accepts a sequence of values.
func (s *SeqChan[T]) Send(seq iter.Seq[T]) {
	if v := s.load(); v != nil {
		safe.Send(func() {
			for msg := range seq {
				v <- msg
			}
		})
	}
}

// TrySend attempts to send a sequence of values on the stored channel.
func (s *SeqChan[T]) TrySend(seq iter.Seq[T]) (sent bool) {
	if v := s.load(); v != nil {
		safe.Send(func() {
			t := time.NewTimer(sendTimeout)
			defer t.Stop()
			for msg := range seq {
				select {
				case v <- msg:
					t.Reset(sendTimeout)
				case <-t.C:
					return
				}
			}
			sent = true
		})
		return
	}
	return
}

// RecvSeq reads values from the stored channel and returns as an iterator.
func (s *SeqChan[T]) Recv() iter.Seq[T] {
	return func(yield func(msg T) bool) {
		for msg := range s.Chan.Recv() {
			if !yield(msg) {
				return
			}
		}
	}
}
