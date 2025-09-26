package errors_test

import (
	"cmp"
	"fmt"
	"strings"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/stack"
)

func getError() error {
	return getStackError()
}

func getStackError() error {
	return errors.Stack(fmt.Errorf("is a stack error"))
}

func TestStackError(t *testing.T) {
	err, _ := errors.As[errors.StackError](getError())
	assert.Error(t, "is a stack error", err)
	assert.Equal(t,
		[]string{
			"go.chrisrx.dev/x/errors_test.getStackError",
			"go.chrisrx.dev/x/errors_test.getError",
			"go.chrisrx.dev/x/errors_test.TestStackError",
		},
		slices.Filter(slices.Map(err.Trace(), func(f stack.Frame) string {
			return f.Name()
		}), func(name string) bool {
			return !cmp.Or(
				strings.HasPrefix(name, "testing"),
				strings.HasPrefix(name, "runtime"),
			)
		}))
}
