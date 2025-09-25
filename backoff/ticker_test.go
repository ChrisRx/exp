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
		ticks := getExpectedDurations(Backoff{}, 4)
		if !ticker.DisableInstantTick {
			ticks = slices.Insert(ticks, 0, 0)
		}
		for _, d := range ticks {
			next := <-ticker.Next()
			assert.WithinDuration(t, last.Add(d), next, 50*time.Millisecond)
			last = next
		}
	})
}
