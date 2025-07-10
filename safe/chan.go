package safe

import (
	"go.chrisrx.dev/x/must"
)

// Close safely closes a Go channel.
func Close[T any, C ~chan T](ch C) {
	if ch == nil {
		return
	}
	defer must.Recover()
	close(ch)
}
