package sync

import (
	"sync"
)

// LazyChan wraps a [Chan], initializing lazily upon initial access.
type LazyChan[T any] struct {
	Chan[T]

	once sync.Once
}

// Close closes the current channel, if present. The stored value is replaced
// with a nil.
//
// If the underlying channel hasn't been initialized yet, then one will be
// created and stored before ultimately being closed. This behavior is
// necessary since the usage of [LazyChan] is predicated upon the underlying
// channel being initialized upon first usage, even if that usage is being
// closed.
func (l *LazyChan[T]) Close() {
	l.init()
	l.Chan.Close()
}

// Load atomically loads the stored channel.
//
// If no channel is currently stored, one will be created and stored.
func (l *LazyChan[T]) Load() chan T {
	l.init()
	return l.Chan.Load()
}

func (l *LazyChan[T]) init() {
	l.once.Do(func() {
		l.Reset()
	})
}
