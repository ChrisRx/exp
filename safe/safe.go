package safe

import (
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
)

func Do(fn func()) (err error) {
	defer must.Catch(ptr.To(err))
	fn()
	return
}
