package must_test

import (
	"testing"

	"go.chrisrx.dev/x/must"
)

func TestClose(t *testing.T) {
	ch := make(chan error)

	must.Close(ch)
	must.Close(ch)
}
