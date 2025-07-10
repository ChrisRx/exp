package must

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

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

func ignoreSource(s stack.Source) bool {
	return Any(
		strings.HasPrefix(s.FullName, "runtime"),
		strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/must"),
		strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/safe"),
	)
}

// TODO: move to a new package
func All[T comparable](S ...T) bool {
	var zero T
	return !slices.Contains(S, zero)
}

// TODO: move to a new package
func Any[T comparable](S ...T) bool {
	var zero T
	for _, elem := range S {
		if elem != zero {
			return true
		}
	}
	return false
}
