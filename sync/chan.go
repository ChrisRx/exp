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
}

// Cap returns the current channel capacity.
func (ch *Chan[T]) Cap() int {
	return cap(ch.Load())
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
//
// If capacity is set to zero, then the channel is unbuffered. If capacity is
// greater than zero, then a buffered channel with the given capacity will be
// created.
func (ch *Chan[T]) New(capacity int) chan T {
	must.Close(ch.swap(make(chan T, capacity)))
	return ch.Load()
}

// Load loads the stored channel. If no channel is stored, a closed channel is
// returned.
func (ch *Chan[T]) Load() chan T {
	if ch := ptr.From(ch.v.Load()); ch != nil {
		return ch
	}
	return makeClosedChannel[T]()
}

// LoadOrNew loads the stored channel, if present. If not present, a new
// channel will be created and the newly stored channel is returned.
func (ch *Chan[T]) LoadOrNew() (_ chan T, isNew bool) {
	const maxAttempts = 100
	for range maxAttempts {
		old := ch.v.Load()
		if v := ptr.From(old); v != nil {
			return v, false
		}
		new := make(chan T, ch.Cap())
		if ch.v.CompareAndSwap(old, &new) {
			return new, true
		}
	}
	panic("Chan.LoadOrNew: infinite loop detected")
}

// Reset constructs and stores a new channel. It is the same as calling [New]
// with the current capacity value.
//
// This is useful when needing to signal that channel consumers should close,
// while establishing a new channel for immediate use.
func (ch *Chan[T]) Reset() chan T {
	return ch.New(ch.Cap())
}

func (ch *Chan[T]) swap(new chan T) (_ chan T) {
	return ptr.From(ch.v.Swap(&new))
}

// makeClosedChannel constructs a new channel of type T that is returned
// closed. This is used to create a new channel that doesn't block.
func makeClosedChannel[T any]() chan T {
	ch := make(chan T)
	defer close(ch)
	return ch
}

// Note that it must not be embedded, due to the Lock and Unlock methods.
// noCopy prevents a struct from being copied after the first use. It achieves
// this by implementing the [sync.Locker] interface, which triggers the go vet
// copylocks check. It should not be embedded.
//
// https://golang.org/issues/8005#issuecomment-190753527
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
