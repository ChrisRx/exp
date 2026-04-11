package group

import (
	"context"
	"iter"

	"go.chrisrx.dev/x/future"
	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/sync"
)

// ResultGroup manages a pool of goroutines that return a result value.
type ResultGroup[T any] struct {
	g, r    *Group
	results sync.Chan[future.Value[T]]
}

// NewResultGroup constructs a new result group using the provided options.
func NewResultGroup[T any](ctx context.Context, opts ...GroupOption) *ResultGroup[T] {
	o := newOptions().Apply(opts)
	r := &ResultGroup[T]{
		g:       New(ctx, opts...),
		r:       New(ctx, opts...),
		results: ptr.From(sync.NewChan[future.Value[T]](o.ResultsBuffer)),
	}
	return r
}

// Go runs the provided function in a goroutine and returns a future containing
// a result value or error.
//
// If an error is encountered, the context for the group is canceled. This
// happens regardless if the error is checked on the future.
//
// If a concurrency limit is set, calls to Go will block once the number of
// running goroutines is reached and will continue blocking until a running
// goroutine returns.
func (r *ResultGroup[T]) Go(fn func(context.Context) (T, error)) future.Value[T] {
	v := future.New(func() (T, error) {
		return fn(r.g.ctx)
	})
	r.g.Go(func(ctx context.Context) error {
		return v.Err()
	})
	r.r.Go(func(ctx context.Context) error {
		select {
		case <-v.Done():
			r.results.Send(v)
		case <-r.r.ctx.Done():
		}
		return nil
	})
	return v
}

// Get returns an iterator of result/error pairs. It blocks until all results
// are read or the group context is done.
func (r *ResultGroup[T]) Get() iter.Seq2[T, error] {
	go func() {
		defer r.results.Reset()
		r.g.Wait()
		r.r.Wait()
	}()

	return func(yield func(T, error) bool) {
		for result := range r.results.Recv() {
			if !yield(result.Get()) {
				return
			}
		}
	}
}

// Collect collects values from the result group until all results are read or
// the group context is done.
func (r *ResultGroup[T]) Collect() ([]T, error) {
	var results []T
	for v, err := range r.Get() {
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

// Wait blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned. It will also
// block until at least one goroutine is scheduled.
func (r *ResultGroup[T]) Wait() error {
	return r.g.Wait()
}
