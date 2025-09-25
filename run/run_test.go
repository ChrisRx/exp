package run_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/run"
)

func TestEvery(t *testing.T) {
	t.Run("every", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		var n int
		run.Every(ctx, func() {
			n++
		}, 100*time.Millisecond)

		assert.Equal(t, 5, n)
	})
}

func TestUntil(t *testing.T) {
	t.Run("until count", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var n int
		run.Until(ctx, func() bool {
			n++
			return n >= 5
		}, 10*time.Millisecond)

		assert.Equal(t, 5, n)
	})

	t.Run("until func()", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var n int
		run.Until(ctx, func() {
			n++
			if n >= 5 {
				cancel()
			}
		}, 10*time.Millisecond)

		assert.Equal(t, 5, n)
	})

	t.Run("until func() bool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var n int
		run.Until(ctx, func() bool {
			n++
			return n >= 5
		}, 10*time.Millisecond)

		assert.Equal(t, 5, n)
	})

	t.Run("until func() error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var n int
		run.Until(ctx, func() error {
			n++
			if n >= 5 {
				return nil
			}
			return fmt.Errorf("retry")
		}, 10*time.Millisecond)

		assert.Equal(t, 5, n)
	})
}
