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
}
