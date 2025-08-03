package must_test

import (
	"errors"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/must"
)

func TestCatch(t *testing.T) {
	expected := errors.New("caught error")
	assert.Error(t, expected, func() (err error) {
		defer must.Catch(&err)
		panic(expected)
	}())
}
