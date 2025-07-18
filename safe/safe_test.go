package safe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/safe"
)

var errPanic = errors.New("do panic")

func TestDo(t *testing.T) {
	assert.ErrorIs(t, safe.Do(func() {}), nil)
	err := safe.Do(func() {
		panic(errPanic)
	})
	assert.ErrorIs(t, err, errPanic)
}
