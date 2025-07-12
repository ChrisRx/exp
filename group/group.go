// Package group provides [group.Group] for managing pools of goroutines.
package group

import (
	"context"

	"go.chrisrx.dev/x/sync"
)

// Group manages a pool of goroutines.
type Group struct {
	wg sync.BoundedWaitGroup

	// The original context provided to [Group.New]. It is needed to reset after
	// calls to [Group.Wait].
	parent context.Context

	// A child context is used internally to cancel spawned goroutines. A new
	// child context is derived after successful calls to [Group.Wait] to allow
	// [Group] to be re-used.
	ctx    context.Context
	cancel context.CancelCauseFunc

	done  sync.Chan[error]
	ready sync.Waiter

	once sync.OnceAgain
	err  error
}

// New constructs a new group using the provided options.
func New(ctx context.Context, opts ...GroupOption) *Group {
	o := newOptions().Apply(opts)
	g := &Group{
		parent: ctx,
	}
	if o.Limit != 0 {
		g.wg.SetLimit(o.Limit)
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
	go func() {
		defer g.ready.Done()
		defer g.wg.Done()

		if err := fn(g.ctx); err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel(g.err)
			})
		}
	}()
	return g
}

// Wait blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered will be returned. It will also
// block until at least one goroutine is scheduled.
func (g *Group) Wait() error {
	defer g.reset()
	g.ready.Wait()
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}

// Done blocks until all the goroutines in this group have returned. If any
// errors occur, the first error encountered is sent on the returned channel,
// otherwise the channel is closed.
func (g *Group) Done() <-chan error {
	done, isNew := g.done.LoadOrNew()
	if isNew {
		go func() {
			defer g.done.Close()
			done <- g.Wait()
		}()
	}
	return done
}

func (g *Group) reset() {
	g.ctx, g.cancel = context.WithCancelCause(g.parent)
	g.once.Reset()
	g.ready.Reset()
}
