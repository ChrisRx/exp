package safe

import (
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/must"
)

// Close safely closes a Go channel.
func Close[T any, C ~chan T](ch C) (closed bool) {
	if ch == nil {
		return
	}
	defer must.Recover(
		errors.RuntimeError("close of closed channel"),
	)
	close(ch)
	return true
}

// Send safely sends on a Go channel.
func Send[T any](ch chan<- T, messages ...T) {
	defer must.Recover(
		errors.RuntimeError("send on closed channel"),
	)
	for _, msg := range messages {
		ch <- msg
	}
}
