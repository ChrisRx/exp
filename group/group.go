// Package group provides types for managing pools of goroutines.
package group

import (
	"context"
	"sync/atomic"
	"time"

	"go.chrisrx.dev/x/run"
	"go.chrisrx.dev/x/sync"
)

// Group manages a pool of goroutines.
type Group struct {
	parent context.Context
	ctx    context.Context
	cancel context.CancelCauseFunc

	limit sync.Semaphore
	wg    sync.WaitGroup
	done  sync.Chan[error]
	ready sync.Waiter

	called atomic.Uint64

	once sync.Once
	err  error
}

// New constructs a new group using the provided options.
func New(ctx context.Context, opts ...GroupOption) *Group {
	o := newOptions().Apply(opts)
	g := &Group{
		parent: ctx,
	}
	if o.Limit != 0 {
		g.limit.SetLimit(o.Limit)
	}
	g.ctx, g.cancel = context.WithCancelCause(ctx)
	return g
}

// Go runs the provided function in a goroutine. If an error is encountered,
// the context for the group is canceled.
//
// If a concurrency limit is set, calls to Go will block once the number of
// running goroutines is reached and will continue blocking until a running
// goroutine returns.
func (g *Group) Go(fn func(context.Context) error) *Group {
	g.wg.Add(1)
	g.called.Add(1)
	g.ready.Done()
	go func() {
		g.limit.Acquire(1)
		defer g.limit.Release()
		defer g.wg.Done()

		// If the context was canceled while waiting to acquire, we shouldn't
		// attempt to run the user-provided function.
		if g.ctx.Err() != nil {
			return
		}

		// The group has encountered an error so user-provided functions shouldn't
		// continue to be executed. This ensures that producers don't needlessly
		// continue performing work when the group has already failed.
		if g.err != nil {
			return
		}

		if err := fn(g.ctx); err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel(g.err)
			})
		}
	}()
	return g
}

// Err returns the last error encountered.
func (g *Group) Err() error {
	return g.err
}

// MustWait blocks until at least n goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned.
//
// MustWait will block until at least one goroutine is scheduled. This is helpful
// to ensure that if Wait is called before calls scheduling goroutines, it
// won't finish waiting prematurely. If there is a possibility that no
// goroutines might be scheduled, then [Wait] should be called instead.
func (g *Group) WaitN(n int) (reterr error) {
	defer g.cancel(reterr)
	defer g.reset()
	return run.Until(g.ctx, func() (bool, error) {
		g.wg.Wait()
		return g.called.Load() >= uint64(n), g.err
	}, run.Options{
		InitialInterval: 10 * time.Millisecond,
		MaxAttempts:     100,
	})
}

// Wait blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned.
//
// Unlike [MustWait], if no goroutines are scheduled, this returns immediately.
func (g *Group) Wait() error {
	defer g.reset()
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}

func (g *Group) reset() {
	g.ctx, g.cancel = context.WithCancelCause(g.parent)
	g.once.Reset()
	g.ready.Reset()
	g.called.Store(0)
}

// Done blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered is sent on the returned channel,
// otherwise the channel is closed.
func (g *Group) Done() <-chan error {
	done, isNew := g.done.LoadOrNew()
	if isNew {
		go func() {
			defer g.done.Close()
			if err := g.Wait(); err != nil {
				done <- err
			}
		}()
	}
	return done
}
