package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackoff(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		var b Backoff
		for _, d := range getExpectedDurations(b, 100_000) {
			assert.Equal(t, d, b.Next())
		}
	})

	t.Run("invalid configuration", func(t *testing.T) {
		cases := []Backoff{
			{MinInterval: -1},
			{MaxInterval: -1},
			{Multiplier: -1},
		}
		for _, tc := range cases {
			assert.PanicsWithValue(t, "cannot provide negative values for backoff", func() {
				tc.Next()
			})
		}
	})
}

func getExpectedDurations(b Backoff, n int) (expected []time.Duration) {
	b.init()
	cur := b.MinInterval
	for range n {
		expected = append(expected, cur)
		if time.Duration(float64(cur)*b.Multiplier) > b.MaxInterval {
			cur = b.MaxInterval
			continue
		}
		cur = time.Duration(float64(cur) * b.Multiplier)
	}
	return expected
}
