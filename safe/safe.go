package safe

import (
	"go.chrisrx.dev/x/must"
)

func Do(fn func()) (err error) {
	defer must.Catch(&err)
	fn()
	return
}
