package run_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/run"
)

func TestRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("retry range", func(t *testing.T) {
		maxAttempts := 2

		var n int
		run.Retry(ctx, func() error {
			return fmt.Errorf("continue")
		}, run.RetryOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxAttempts:     maxAttempts,
		}).Range(func(attempt int, err error) {
			n++
		})
		assert.Equal(t, maxAttempts, n)
	})

	t.Run("retry iter", func(t *testing.T) {
		maxAttempts := 5

		var n int
		for attempts, err := range run.Retry(ctx, func() error {
			return fmt.Errorf("continue")
		}, run.RetryOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxAttempts:     5,
		}) {
			slog.Default().Info("attempt failed",
				slog.Any("error", err),
				slog.Int("attempts", attempts),
			)
			n++
		}
		assert.Equal(t, maxAttempts, n)
	})

	t.Run("recovers from panic", func(t *testing.T) {
		maxAttempts := 5

		var n int
		for attempts, err := range run.Retry(ctx, func() error {
			panic("something real bad happened")
		}, run.RetryOptions{
			InitialInterval: 10 * time.Millisecond,
			MaxAttempts:     5,
		}) {
			slog.Default().Info("attempt failed",
				slog.Any("error", err),
				slog.Int("attempts", attempts),
			)
			n++
		}
		assert.Equal(t, maxAttempts, n)
	})
}
