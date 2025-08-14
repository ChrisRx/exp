package sync

import (
	"sync/atomic"
	"time"

	"go.chrisrx.dev/x/chans"
	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/safe"
)

// A Chan is a channel of type T, synchronized with lock-free atomic
// operations.
//
// The initial value is nil, requiring a call to a method that initializes a
// channel, such as [Chan.New] or [Chan.Reset]. If lazy initialization is
// needed, [LazyChan] can be used.
//
// Since Chan uses [atomic.Pointer] to store the underlying channel, it
// inherits the same restrictions disallowing copying and spurious type
// conversion.
type Chan[T any] struct {
	v atomic.Pointer[chan T]
}

// NewChan constructs a new [Chan] of type T. If capacity is greater than zero,
// it is initialized with a buffered channel, otherwise the channel is
// unbuffered.
func NewChan[T any](capacity int) *Chan[T] {
	var ch Chan[T]
	ch.New(capacity)
	return &ch
}

// Cap returns the current channel capacity.
func (ch *Chan[T]) Cap() int {
	return cap(ch.Load())
}

// Close closes the current channel, if present. The stored value is replaced
// with a nil.
func (ch *Chan[T]) Close() {
	safe.Close(ch.swap(nil))
}

// Closed returns true when the current channel is closed.
func (ch *Chan[T]) Closed() bool {
	return ch.load() == nil
}

// New constructs and stores a new channel. If a channel is already stored, it
// is closed after being replaced.
//
// If capacity is set to zero, then the channel is unbuffered. If capacity is
// greater than zero, then a buffered channel with the given capacity will be
// created.
func (ch *Chan[T]) New(capacity int) chan T {
	safe.Close(ch.swap(make(chan T, capacity)))
	return ch.Load()
}

// Load loads the stored channel. If no channel is stored, a closed channel is
// returned.
func (ch *Chan[T]) Load() chan T {
	if ch := ch.load(); ch != nil {
		return ch
	}
	return makeClosedChannel[T]()
}

// makeClosedChannel constructs a new channel of type T that is returned
// closed. This is used to create a new channel that doesn't block.
func makeClosedChannel[T any]() chan T {
	ch := make(chan T)
	defer close(ch)
	return ch
}

// LoadOrNew loads the stored channel, if present. If not present, a new
// channel will be created and the newly stored channel is returned.
func (ch *Chan[T]) LoadOrNew() (_ chan T, isNew bool) {
	const maxAttempts = 1000
	for range maxAttempts {
		old := ch.v.Load()
		if v := ptr.From(old); v != nil {
			return v, false
		}
		new := make(chan T, ch.Cap())
		if ch.v.CompareAndSwap(old, &new) {
			safe.Close(ptr.From(old))
			return new, true
		}
	}
	// This will only be reached if some kind of bug was introduced into the
	// loop.
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

// Recv returns the stored channel as receive-only.
func (ch *Chan[T]) Recv() <-chan T {
	return ch.Load()
}

// CloseAndRecv closes the stored channel and returns a results channel with
// the remaining elements.
func (ch *Chan[T]) CloseAndRecv() <-chan T {
	if v := ch.load(); v != nil {
		return chans.Drain(v)
	}
	return makeClosedChannel[T]()
}

// Send sends the provided messages on the stored channel. The messages are
// sent sequentially in the order they are provided.
//
// Send has the same behavior expected as directly using a Go channel. If the
// stored channel is unbuffered, calls to Send will block until a reader is
// receives the message. A buffered channel will send immedidately up to the
// available capacity.
//
// If the stored channel is uninitialized or closed, it returns immediately.
func (ch *Chan[T]) Send(messages ...T) {
	if v := ch.load(); v != nil {
		safe.Send(func() {
			for _, msg := range messages {
				v <- msg
			}
		})
	}
}

const sendTimeout = 100 * time.Millisecond

// TrySend attempts to send a value on the stored channel. The messages are
// sent sequentially in the order they are provided.
//
// Unlike [Chan.Send], it will wait for the value to be sent for 100ms before
// returning. If the channel is closed while attempting to send a value, the
// send on closed panic is recovered and logged.
//
// If the stored channel is uninitialized or closed, it returns immediately.
func (ch *Chan[T]) TrySend(messages ...T) (sent bool) {
	if v := ch.load(); v != nil {
		safe.Send(func() {
			t := time.NewTimer(sendTimeout)
			defer t.Stop()
			for _, msg := range messages {
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

func (ch *Chan[T]) load() chan T {
	return ptr.From(ch.v.Load())
}

func (ch *Chan[T]) swap(new chan T) chan T {
	return ptr.From(ch.v.Swap(&new))
}
