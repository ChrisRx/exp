package must

import (
	"fmt"
	"log/slog"
	"strings"

	"go.chrisrx.dev/x/cmp"
	"go.chrisrx.dev/x/stack"
)

func Recover() {
	if r := recover(); r != nil {
		slog.Error("panic",
			slog.String("loc", stack.GetLocation(ignoreSource)),
			slog.Any("err", r),
		)
	}
}

func ignoreSource(s stack.Source) bool {
	return cmp.Any(
		strings.HasPrefix(s.FullName, "runtime"),
		strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/must"),
		strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/safe"),
	)
}

func Catch(err *error) {
	if r := recover(); r != nil {
		switch t := r.(type) {
		case error:
			*err = fmt.Errorf("panic: %w", t)
		default:
			*err = fmt.Errorf("panic: %v", t)
		}
	}
}
