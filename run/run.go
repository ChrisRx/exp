package run

import (
	"context"
	"log/slog"
	"time"

	"go.chrisrx.dev/x/log"
)

func Every(ctx context.Context, fn func() error, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := fn(); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func Until(ctx context.Context, fn func() error, interval time.Duration) error {
	logger := log.FromContext(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := fn(); err != nil {
		logger.Debug("encountered error while looping", slog.Any("error", err))
	} else {
		return nil
	}

	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				logger.Debug("encountered error while looping", slog.Any("error", err))
				continue
			}
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}
