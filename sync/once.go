package sync

import (
	"sync/atomic"

	"go.chrisrx.dev/x/safe"
)

// Once is a resettable version of [sync.Once].
type Once struct {
	_ safe.NoCopy

	done atomic.Bool
	m    Mutex
}

func (o *Once) Do(f func()) {
	if !o.done.Load() {
		o.doSlow(f)
	}
}

// Reset resets the finished state, allowing reuse.
func (o *Once) Reset() {
	o.done.Store(false)
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.done.Load() {
		defer o.done.Store(true)
		f()
	}
}
