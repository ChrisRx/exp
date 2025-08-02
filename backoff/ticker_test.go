package backoff

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

func TestTicker(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var ticker Ticker
		last := time.Now()
		for _, d := range getExpectedDurations(Backoff{}, 4) {
			next := <-ticker.Next()
			assert.WithinDuration(t, last.Add(d), next, 10*time.Millisecond)
			last = next
		}
	})
}
