package assert

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"text/template"
	"time"
)

func Equal[T any](tb testing.TB, expected, actual T, args ...any) {
	tb.Helper()

	if !equal(expected, actual) {
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "not equal"),
			Expected: print(expected),
			Actual:   print(actual),
		})
	}
}

func Panic(tb testing.TB, expected any, fn func(), args ...any) {
	tb.Helper()

	r := func(fn func()) (retr any) {
		defer func() { retr = recover() }()
		fn()
		return
	}(fn)

	if expected != nil && r == nil {
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "expected panic"),
			Expected: expected,
		})
	}
	if !equal(expected, r) {
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "unexpected panic"),
			Expected: expected,
			Actual:   r,
		})
	}
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
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "unexpected error"),
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
		fail(tb, Message{
			Message: cmp.Or(getMessage(args...), "unexpected error"),
			Expected: strings.Join(Map(unwrap(expected), func(elem error) string {
				return elem.Error()
			}), "\n\t"),
			Actual: actual,
		})
		return false
	case string:
		m, err := newMatcher(expected)
		if err != nil {
			tb.Fatal(err)
			return false
		}
		tb.Helper()
		if m(actual.Error()) {
			return true
		}
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "unexpected error"),
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
	if len(missing) == 0 && len(extra) == 0 {
		return true
	}

	fail(tb, Message{
		Message: cmp.Or(getMessage(args...), "elements do not match"),
		Elements: slices.Concat(
			Map(missing, func(elem T) any {
				return fmt.Sprintf("- (%[1]T)(%[1]v)", elem)
			}),
			Map(extra, func(elem T) any {
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

	if compare(actual, start) == -1 || compare(actual, end) == 1 {
		tb.Helper()
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "not between"),
			Expected: fmt.Sprintf("%s <-> %v", format(start), format(end)),
			Actual:   format(actual),
		})
	}
}

func WithinDuration(tb testing.TB, expected, actual time.Time, delta time.Duration, args ...any) {
	tb.Helper()

	d := expected.Sub(actual)
	if d.Abs() > delta {
		fail(tb, Message{
			Message:  cmp.Or(getMessage(args...), "not within expected delta"),
			Expected: fmt.Sprintf("%v maxdelta=%v", format(expected), delta),
			Actual:   fmt.Sprintf("%v delta=%v", format(actual), d.Abs()),
		})
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

func format(v any) string {
	switch t := v.(type) {
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

var failureMessage = template.Must(template.New("").Funcs(map[string]any{
	"indent": func(spaces int, v any) string {
		pad := strings.Repeat(" ", spaces)
		return pad + strings.ReplaceAll(fmt.Sprint(v), "\n", "\n"+pad)
	},
}).Parse(`
{{- .Message }}
{{- with .Expected }}
expected:
{{ . | indent 4 }}
{{- end -}}
{{- with .Actual }}
actual:
{{ . | indent 4 }}
{{- end -}}
{{- range .Elements }}
	{{ . }}
{{- end -}}
`))

type Message struct {
	Message  string
	Expected any
	Actual   any
	Elements []any
}

func fail(tb testing.TB, m Message) {
	var b bytes.Buffer
	if err := failureMessage.Execute(&b, m); err != nil {
		panic(err)
	}
	tb.Helper()
	tb.Fatal(b.String())
}
