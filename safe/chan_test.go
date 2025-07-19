package safe_test

import (
	"testing"

	"go.chrisrx.dev/x/safe"
)

func TestClose(t *testing.T) {
	ch := make(chan error)

	safe.Close(ch)
	safe.Close(ch)
}
