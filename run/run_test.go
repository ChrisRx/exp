package run_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.chrisrx.dev/x/run"
)

func TestEvery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var n int
	run.Every(ctx, func() {
		n++
	}, 100*time.Millisecond)

	assert.Equal(t, 5, n)
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
}
