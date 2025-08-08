// Package future provides types for representing values that are not
// immediately ready.
package future

// Value is a value that may not yet be ready.
type Value[T any] interface {
	// Get blocks until the result function is complete.
	Get() (T, error)
}

type future[T any] struct {
	done chan struct{}
	fn   func() (T, error)

	value T
	err   error
}

// New constructs a new future value using the provided result function. The
// result function is called immediately in a new goroutine.
func New[T any](fn func() (T, error)) Value[T] {
	f := &future[T]{
		done: make(chan struct{}),
		fn:   fn,
	}
	go func() {
		defer close(f.done)
		f.value, f.err = f.fn()
	}()
	return f
}

// Get blocks until result function is complete.
func (f *future[T]) Get() (T, error) {
	<-f.done
	return f.value, f.err
}
