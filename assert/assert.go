package assert

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert/internal/slices"
)

func Equal[T any](tb testing.TB, expected, actual T, args ...any) {
	if equal(expected, actual) {
		return
	}

	tb.Helper()
	Fatal(tb, Message{
		Header: header("not equal", args),
		Diff:   Diff(expected, actual),
	})
}

func Panic(tb testing.TB, expected any, fn func(), args ...any) {
	r := func() (retr any) {
		defer func() { retr = recover() }()
		fn()
		return
	}()

	if !isZero(expected) && r == nil {
		tb.Helper()
		Fatal(tb, Message{
			Header:   header("expected panic", args),
			Expected: expected,
		})
	}

	if equal(expected, r) {
		return
	}

	tb.Helper()
	Fatal(tb, Message{
		Header:   header("unexpected panic", args),
		Expected: cmp.Or(expected, "<nil>"),
		Actual:   r,
	})
}

func NoPanic(tb testing.TB, fn func(), args ...any) {
	tb.Helper()
	Panic(tb, nil, fn, args...)
}

func Error(tb testing.TB, expected any, actual error, args ...any) bool {
	if expected == nil {
		if errors.Is(nil, actual) {
			return true
		}

		tb.Helper()
		Fatal(tb, Message{
			Header:   header("unexpected error", args),
			Expected: expected,
			Actual:   actual,
		})
		return false
	}

	switch expected := expected.(type) {
	case error:
		if errors.Is(expected, actual) {
			return true
		}

		tb.Helper()
		Fatal(tb, Message{
			Header: header("unexpected error", args),
			Expected: strings.Join(slices.Map(unwrap(expected), func(elem error) string {
				return elem.Error()
			}), "\n"),
			Actual: actual,
		})
		return false
	case string:
		fn, err := newMatcherFunc(expected)
		if err != nil {
			tb.Fatal(err)
			return false
		}

		tb.Helper()
		if fn(actual.Error()) {
			return true
		}
		Fatal(tb, Message{
			Header:   header("unexpected error", args),
			Expected: expected,
			Actual:   actual,
		})
		return false
	default:
		tb.Fatalf("received invalid type: %T", expected)
		return false
	}
}

func NoError(tb testing.TB, actual error, args ...any) bool {
	tb.Helper()
	return Error(tb, nil, actual, args...)
}

func ElementsMatch[T any](tb testing.TB, expected, actual []T, args ...any) bool {
	if len(expected) == 0 && len(actual) == 0 {
		return true
	}

	missing := slices.Filter(expected, func(elem T) bool {
		return !contains(actual, elem)
	})
	added := slices.Filter(actual, func(elem T) bool {
		return !contains(expected, elem)
	})

	if len(missing) == 0 && len(added) == 0 {
		return true
	}

	tb.Helper()
	Fatal(tb, Message{
		Header: header("elements do not match", args),
		Elements: slices.Concat(
			slices.Map(missing, func(elem T) any {
				return fmt.Sprintf("- (%[1]T)(%[1]v)", elem)
			}),
			slices.Map(added, func(elem T) any {
				return fmt.Sprintf("+ (%[1]T)(%[1]v)", elem)
			}),
		),
	})
	return false
}

func Between(tb testing.TB, start, end, actual any, args ...any) {
	if !isComparable(actual) {
		tb.Fatalf("received incomparable type: %T", actual)
	}

	if compare(actual, start) >= 0 && compare(actual, end) <= 0 {
		return
	}

	tb.Helper()
	Fatal(tb, Message{
		Header:   header("not between", args),
		Expected: fmt.Sprintf("%s <-> %v", format(start), format(end)),
		Actual:   format(actual),
	})
}

func WithinDuration(tb testing.TB, expected, actual time.Time, delta time.Duration, args ...any) {
	if expected.Sub(actual).Abs() <= delta {
		return
	}

	tb.Helper()
	Fatal(tb, Message{
		Header:   header("not within expected delta", args),
		Expected: fmt.Sprintf("%v ±%v", format(expected), delta),
		Actual:   fmt.Sprintf("%v %v", format(actual), format(actual.Sub(expected))),
	})
}

func header(defaultMsg string, args []any) string {
	if len(args) == 0 {
		return defaultMsg
	}
	format, ok := args[0].(string)
	if !ok {
		panic(fmt.Errorf("expected string, received %T", args[0]))
	}
	return fmt.Sprintf(format, args[1:]...)
}

func format(v any) string {
	switch t := v.(type) {
	case time.Duration:
		sign := "-"
		if t > 0 {
			sign = "+"
		}
		return fmt.Sprintf("%s%v", sign, v)
	case time.Time:
		return t.Format("2006-01-02T15:04:05.000Z")
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(v)
	}
}

func unwrap(err error) (errs []error) {
	errs = append(errs, err)
	switch x := err.(type) {
	case interface{ Unwrap() error }:
		if err := x.Unwrap(); err != nil {
			errs = append(errs, unwrap(err)...)
		}
		return errs
	case interface{ Unwrap() []error }:
		for _, err := range x.Unwrap() {
			errs = append(errs, unwrap(err)...)
		}
		return errs
	default:
		return errs
	}
}
