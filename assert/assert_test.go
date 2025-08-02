package assert_test

import (
	"errors"
	"testing"

	"go.chrisrx.dev/x/assert"
)

func TestAssert(t *testing.T) {
	assert.Equal(t, "testing", "testing")
	assert.Panics(t, "send on closed channel", func() {
		ch := make(chan struct{})
		close(ch)
		ch <- struct{}{}
	})
	err := errors.New("some error")
	assert.ErrorIs(t, err, func() error {
		return err
	}())
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, []int{5, 3, 2, 1, 4})
}
