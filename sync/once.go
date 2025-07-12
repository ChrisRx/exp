package sync

import (
	"sync/atomic"

	"go.chrisrx.dev/x/safe"
)

// OnceAgain is a resettable version of [sync.Once].
type OnceAgain struct {
	_ safe.NoCopy

	done atomic.Bool
	m    Mutex
}

func (o *OnceAgain) Do(f func()) {
	if o.done.Load() {
		o.doSlow(f)
	}
}

// Reset resets the finished state, allowing reuse.
func (o *OnceAgain) Reset() {
	o.done.Store(false)
}

func (o *OnceAgain) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done.Load() {
		defer o.done.Store(true)
		f()
	}
}
