package safe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/safe"
)

func TestClose(t *testing.T) {
	ch := make(chan error)

	safe.Close(ch)
	safe.Close(ch)
}

func TestCatch(t *testing.T) {
	expected := errors.New("caught error")
	assert.Error(t, expected, func() (err error) {
		defer must.Catch(ptr.To(err))
		panic(expected)
	}())
}
