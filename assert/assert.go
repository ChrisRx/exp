package assert

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func Equal[T any](tb testing.TB, expected, actual T, args ...any) {
	tb.Helper()

	if !equal(expected, actual) {
		tb.Fatalf("%s\n%s",
			cmp.Or(getMessage(args...), "not equal"),
			Diff(expected, actual),
		)
	}
}

func Panics(tb testing.TB, expected any, fn func(), args ...any) {
	tb.Helper()

	defer func() {
		if r := recover(); r != nil {
			if !equal(expected, r) {
				tb.Fatalf("%s\nexpected:\n\t%v\nactual:\n\t%v\n",
					cmp.Or(getMessage(args...), "unexpected panic"),
					expected,
					r,
				)
			}
		}
	}()

	fn()

	tb.Fatalf("%s\nexpected:\n%v\n",
		cmp.Or(getMessage(args...), "expected panic"),
		expected,
	)
}

func NoPanics(tb testing.TB, fn func(), args ...any) {
	tb.Helper()

	defer func() {
		if r := recover(); r != nil {
			tb.Fatalf("%s\n\t%v\n",
				cmp.Or(getMessage(args...), "unexpected panic"),
				r,
			)
		}
	}()

	fn()
}

func ErrorIs(tb testing.TB, expected, actual error, args ...any) bool {
	tb.Helper()

	if !errors.Is(expected, actual) {
		// TODO(chrism): print error chain
		tb.Fatalf("%s\nexpected:\n\t%v\nactual:\n\t%v\n",
			cmp.Or(getMessage(args...), "unexpected error"),
			expected,
			actual,
		)
		return false
	}
	return true
}

func NoError(tb testing.TB, actual error, args ...any) bool {
	tb.Helper()

	if actual != nil {
		tb.Fatalf("%s\nactual:\n\t%v\n",
			cmp.Or(getMessage(args...), "unexpected error"),
			actual,
		)
		return false
	}
	return true
}

func ElementsMatch[T any](tb testing.TB, expected, actual []T, args ...any) bool {
	tb.Helper()

	if len(expected) == 0 && len(actual) == 0 {
		return true
	}

	missing := make([]T, 0)
	for _, elem := range expected {
		if !contains(actual, elem) {
			missing = append(missing, elem)
		}
	}
	extra := make([]T, 0)
	for _, elem := range actual {
		if !contains(expected, elem) {
			extra = append(extra, elem)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		tb.Fatalf("%s\n\t%v\n\t%v\n",
			cmp.Or(getMessage(args...), "elements do not match"),
			strings.Join(Map(missing, func(elem T) string {
				return fmt.Sprintf("- (%[1]T)(%[1]v)", elem)
			}), "\n\t"),
			strings.Join(Map(extra, func(elem T) string {
				return fmt.Sprintf("+ (%[1]T)(%[1]v)", elem)
			}), "\n\t"),
		)
		return false
	}

	return true
}

func WithinDuration(tb testing.TB, expected, actual time.Time, delta time.Duration, args ...any) {
	tb.Helper()

	d := expected.Sub(actual)
	if d < -delta || d > delta {
		tb.Fatalf("%s\nexpected:\n\t%v\nactual:\n\t%v\ndelta:\n\t%v\nactual delta:\n\t%v\n",
			cmp.Or(getMessage(args...), "not within expected delta"),
			expected,
			actual,
			delta,
			d.Abs(),
		)
	}
}

func getMessage(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	format, ok := args[0].(string)
	if !ok {
		panic(fmt.Errorf("expected string, received %T", args[0]))
	}
	return fmt.Sprintf(format, args[1:]...)
}
