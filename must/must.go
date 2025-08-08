// Package must provides functions for handling errors and panics.
package must

import "log/slog"

// Ok takes a function that returns a value and error and returns only a
// value. If an error is encountered, it is logged using the default logger and
// the zero value of type T is returned.
func Ok[T any](v T, err error) (zero T) {
	if err != nil {
		slog.Error(err.Error())
		return zero
	}
	return v
}

func Get0[T1, T2 any](v1 T1, v2 T2) T1 { return v1 }
func Get1[T1, T2 any](v1 T1, v2 T2) T2 { return v2 }
