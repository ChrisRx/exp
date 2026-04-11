// Package future provides types for representing values that are not
// immediately ready.
package future

import (
	"go.chrisrx.dev/x/safe"
	"go.chrisrx.dev/x/sync"
)

// Value is a value that may not yet be ready.
type Value[T any] interface {
	// Get blocks until the result function is complete.
	Get() (T, error)
	Err() error
	Done() <-chan struct{}
}

type future[T any] struct {
	ready, done chan struct{}
	fn          func() (T, error)

	once sync.Once

	value T
	err   error
}

// New constructs a new future value using the provided result function. The
// result function will not run until a method on [Future.Value] is called.
func New[T any](fn func() (T, error)) Value[T] {
	f := &future[T]{
		ready: make(chan struct{}),
		done:  make(chan struct{}),
		fn:    fn,
	}
	return f
}

func (f *future[T]) run() {
	f.once.Do(func() {
		go func() {
			defer close(f.ready)
			f.value, f.err = f.fn()
		}()
	})
}

// Get blocks until result function is complete.
func (f *future[T]) Get() (T, error) {
	f.run()
	defer safe.Close(f.done)
	<-f.ready
	return f.value, f.err
}

func (f *future[T]) Err() error {
	f.run()
	defer safe.Close(f.done)
	<-f.ready
	return f.err
}

func (f *future[T]) Done() <-chan struct{} {
	f.run()
	return f.done
}
