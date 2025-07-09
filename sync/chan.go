package sync

import (
	"sync/atomic"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
)

// A Chan is a channel of type T, synchronized with lock-free atomic
// operations.
//
// The initial value is nil, requiring a call to a method that initializes a
// channel, such as [Chan.New] or [Chan.Reset]. If lazy initialization is
// needed, [LazyChan] can be used.
type Chan[T any] struct {
	_ [0]*T  // prevent type casting
	_ noCopy // prevent copying

	v atomic.Pointer[chan T]
	n int
}

// Close closes the current channel, if present. The stored value is replaced
// with a nil.
func (ch *Chan[T]) Close() {
	must.Close(ch.swap(nil))
}

// Closed returns true when the current channel is closed.
func (ch *Chan[T]) Closed() bool {
	return ch.Load() == nil
}

// New constructs and stores a new channel. If a channel is already stored, it
// is closed after being replaced.
func (ch *Chan[T]) New(capacity int) chan T {
	ch.n = capacity
	return ch.Reset()
}

// Load loads the stored channel. If no channel is stored, a closed channel is
// returned.
func (ch *Chan[T]) Load() chan T {
	if ch := ptr.From(ch.v.Load()); ch != nil {
		return ch
	}
	return ClosedChannel[T]()
}

// LoadOrNew loads the stored channel, if present. If not present, a new
// channel will be created and the newly stored channel is returned.
func (ch *Chan[T]) LoadOrNew() (_ chan T, isNew bool) {
	for {
		old := ch.v.Load()
		if v := ptr.From(old); v != nil {
			return v, false
		}
		new := make(chan T, ch.n)
		if ch.v.CompareAndSwap(old, &new) {
			return new, true
		}
	}
}

// Reset replaces the current channel with a new channel and closes the
// previous channel. This is useful when needing to signal that channel
// consumers should close, while establishing a new channel for immediate use.
func (ch *Chan[T]) Reset() chan T {
	must.Close(ch.swap(make(chan T, ch.n)))
	return ch.Load()
}

// Cap returns the current channel capacity.
func (ch *Chan[T]) Cap() int { return ch.n }

// SetCap sets the channel capacity. If capacity is set to zero, then the
// channel is unbuffered. If capacity is greater than zero, then a buffered
// channel with the given capacity will be created.
func (ch *Chan[T]) SetCap(capacity int) {
	ch.New(capacity)
}

func (ch *Chan[T]) swap(new chan T) (_ chan T) {
	return ptr.From(ch.v.Swap(&new))
}

// closedchan is a reusable closed channel.
var closedchan = ClosedChannel[struct{}]()

func ClosedChannel[T any]() chan T {
	ch := make(chan T)
	defer close(ch)
	return ch
}
