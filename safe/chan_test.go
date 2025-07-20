package safe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.chrisrx.dev/x/safe"
)

func TestClose(t *testing.T) {
	ch := make(chan error)
	close(ch)
	safe.Close(ch)
}

func TestSend(t *testing.T) {
	ch := make(chan int)
	close(ch)
	assert.PanicsWithError(t, "send on closed channel", func() {
		ch <- 10
	})
	var sent bool
	safe.Send(func() {
		ch <- 10
		sent = true
	})
	assert.Equal(t, false, sent)
}
