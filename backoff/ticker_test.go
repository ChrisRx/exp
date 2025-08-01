package backoff

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	var ticker Ticker
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)
	defer cancel()

	for {
		select {
		case <-ticker.Next():
		case <-ctx.Done():
			return
		}
	}
}
