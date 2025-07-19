package safe

import (
	"go.chrisrx.dev/x/must"
)

// Do calls the provided function, returning any panic as an error.
func Do(fn func()) (err error) {
	defer must.Catch(&err)
	fn()
	return
}
