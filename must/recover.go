package must

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"go.chrisrx.dev/x/cmp"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/stack"
)

// Recover can be used to recover from panics within a function. It must be
// called with defer where the panic occurs. If errors are provided, Recover
// will only recover from panics that match one of the errors. If no errors are
// provided, then all panics will be recovered and logged using the default
// slog logger.
func Recover(errs ...error) {
	if r := recover(); r != nil {
		if len(errs) == 0 {
			slog.Error("panic",
				slog.String("loc", stack.GetLocation(func(s stack.Source) bool {
					return cmp.Any(
						strings.HasPrefix(s.FullName, "runtime"),
						strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/must"),
						strings.HasPrefix(s.FullName, "go.chrisrx.dev/x/safe"),
					)
				})),
				slog.Any("err", asError(r)),
			)
			return
		}

		switch v := r.(type) {
		case runtime.Error:
			// There are runtime errors that aren't public that can be differentiated
			// by interface. These should be compared by error string value instead
			// of using [errors.Is].
			for _, err := range errs {
				if err.Error() == v.Error() {
					return
				}
			}
		case error:
			for _, err := range errs {
				if errors.Is(v, err) {
					return
				}
			}
		default:
			for _, err := range errs {
				if err.Error() == fmt.Errorf("%v", r).Error() {
					return
				}
			}
		}

		// No errors matched so panicking again with the originally recovered
		// value.
		panic(r)
	}
}

// RecoverFunc allows recovering a panic within a function. If the function
// provided returns true, then the panic is recovered, otherwise panic is
// called with the original recovered value.
func RecoverFunc(fn func(r any) bool) {
	if r := recover(); r != nil {
		if !fn(r) {
			panic(r)
		}
	}
}

func Catch(err *error) {
	if r := recover(); r != nil {
		*err = asError(r)
	}
}

func asError(r any) error {
	switch t := r.(type) {
	case error:
		return t
	default:
		return fmt.Errorf("%v", t)
	}
}
