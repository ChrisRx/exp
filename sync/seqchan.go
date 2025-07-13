package sync

import (
	"iter"
	"time"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
)

// SeqChan wraps a [Chan], providing Send/Recv methods that operate on iterator
// sequences.
type SeqChan[T any] struct {
	Chan[T]
}

// NewBufferedChan constructs a new buffered [Chan] of type T with the provided
// capacity.
func NewSeqChan[T any](capacity int) *SeqChan[T] {
	return &SeqChan[T]{
		Chan: ptr.From(NewChan[T](capacity)),
	}
}

// SendSeq attempts to send a sequence of values on the stored channel.
func (s *SeqChan[T]) Send(seq iter.Seq[T]) (sent bool) {
	if v := s.load(); v != nil {
		t := time.NewTimer(sendTimeout)
		defer t.Stop()

		defer must.Recover()
		for msg := range seq {
			select {
			case v <- msg:
				t.Reset(sendTimeout)
			case <-t.C:
				return false
			}
		}
		return true
	}
	return false
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
