// Package convert provides functions for registering and using functions that
// convert from one Go type to another. This package makes heavy use of
// reflection and the purpose is not designed for high-performance.
package convert

import (
	"fmt"
)

// Into converts the provided value into the return value of a different type,
// using registered conversion functions.
func Into[To, From any](in From, opts ...Option) (To, error) {
	fn, ok := LookupFor[From, To]()
	if !ok {
		var zero To
		return zero, fmt.Errorf("no conversion function found: %T->%T", in, zero)
	}
	return fn(in, opts...)
}
