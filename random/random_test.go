package random

import (
	"math"
	"testing"
	"testing/cryptotest"

	"go.chrisrx.dev/x/assert"
)

func TestString(t *testing.T) {
	t.Run("", func(t *testing.T) {
		cryptotest.SetGlobalRandom(t, 0)

		cases := []struct {
			length   int
			expected string
		}{
			{
				length:   64,
				expected: "vb2kLSeMqN3ycDdJrvD03BuEIzIauoSfKoNhzWVeWmyXhBvLD8Shg94yYJoGR1FI",
			},
		}
		for _, tt := range cases {
			assert.Equal(t, tt.expected, String(tt.length))
		}
	})

	t.Run("", func(t *testing.T) {
		nums := make(map[int]int)
		testOnlyRejectionSampling = func(i int) {
			nums[i]++
		}

		for range 1_000_000 {
			String(64)
		}
		assert.Between(t, 0.0, 0.01, stddev(nums))
	})
}

func stddev(data map[int]int) float64 {
	if len(data) <= 1 {
		return 0.0
	}
	var sum, mean, sd float64
	for _, n := range data {
		sum += float64(n)
	}
	mean = sum / float64(len(data))

	for _, n := range data {
		sd += math.Pow(float64(n)-mean, 2)
	}
	return math.Sqrt(sd/float64(len(data)-1)) / mean
}
