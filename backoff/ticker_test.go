package backoff

import (
	"slices"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

func TestTicker(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var ticker Ticker
		last := time.Now()
		for _, d := range slices.Insert(getExpectedDurations(Backoff{}, 4), 0, 0) {
			next := <-ticker.Next()
			assert.WithinDuration(t, last.Add(d), next, 50*time.Millisecond)
			last = next
		}
	})
}
