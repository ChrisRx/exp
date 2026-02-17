package slices_test

import (
	"strconv"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/slices"
)

func TestSlices(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5}
	result := slices.Map(ints, func(v int) string {
		return strconv.Itoa(v)
	})
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, result)

	t.Run("FilterMap", func(t *testing.T) {
		assert.Equal(t, slices.FilterMap([]string{"a", "", "d", "", "c", "e", "b"}, func(s string) string {
			return s
		}), []string{"a", "d", "c", "e", "b"})
	})

	t.Run("MapEntries", func(t *testing.T) {
		assert.Equal(t, slices.MapEntries([]string{"a", "d", "c", "e", "b"}, func(s string) (string, string) {
			return s, s
		}), map[string]string{"a": "a", "d": "d", "c": "c", "e": "e", "b": "b"})
	})
}
