package group_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/chans"
	"go.chrisrx.dev/x/group"
	"go.chrisrx.dev/x/sync"
)

func TestGroup(t *testing.T) {
	// number of goroutines
	n := 10

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		var done atomic.Uint32
		g := group.New(t.Context())
		for range n {
			g.Go(func(ctx context.Context) error {
				defer done.Add(1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, n, int(done.Load()))
	})

	t.Run("method chaining", func(t *testing.T) {
		t.Parallel()
		var done atomic.Uint32
		if err := group.New(t.Context()).Go(func(ctx context.Context) error {
			defer done.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		}).Go(func(ctx context.Context) error {
			defer done.Add(1)
			time.Sleep(500 * time.Millisecond)
			return nil
		}).Wait(); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, int(done.Load()))
	})

	t.Run("wait again", func(t *testing.T) {
		t.Parallel()
		var done atomic.Uint32
		g := group.New(t.Context())
		for range n {
			g.Go(func(ctx context.Context) error {
				defer done.Add(1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		for range n {
			g.Go(func(ctx context.Context) error {
				defer done.Add(1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, n*2, int(done.Load()))
	})

	t.Run("timeout", func(t *testing.T) {
		t.Parallel()
		var done atomic.Uint32
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := group.New(ctx)
		g.Go(func(ctx context.Context) error {
			defer done.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		for range n {
			g.Go(func(ctx context.Context) error {
				defer done.Add(1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}
		assert.Error(t, context.DeadlineExceeded, g.Wait())
		assert.Equal(t, n+1, int(done.Load()))
	})

	t.Run("multiple wait callers", func(t *testing.T) {
		t.Parallel()
		var done atomic.Uint32
		g := group.New(t.Context())
		for range n {
			g.Go(func(ctx context.Context) error {
				defer done.Add(1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}

		ch := sync.NewChan[error](1)
		defer ch.Close()
		go func() { ch.Send(g.Wait()) }()
		go func() { ch.Send(g.Wait()) }()
		for _, err := range chans.CollectN(ch.Recv(), 2) {
			assert.NoError(t, err)
		}
		assert.Equal(t, n, int(done.Load()))
	})
}
