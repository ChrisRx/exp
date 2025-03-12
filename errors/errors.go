// Package errors is a mostly drop-in replacement for the standard library
// errors package. It provides some extra functions for handling errors, and in
// cases, like [errors.As], supplants the standard library version with a
// function using generics.
package errors

import (
	"errors"
	"fmt"
)

// As is a generic version of the stdlib errors.As. The purpose is to allow for
// better ergonomics, changing this:
//
//	var someErr SomeError
//	if errors.As(err, &someErr) {
//		...
//	}
//
// into this:
//
//	if someErr, ok := xerrors.As[*SomeError](err); ok {
//		...
//	}
func As[T error](err error) (_ T, ok bool) {
	var v T
	if errors.As(err, &v) {
		return v, true
	}
	return v, false
}

// Wrap returns a new error wrapping the provided error with additional
// context.
//
// If the provided error is nil, then nil is returned. This enables using Wrap
// with a return value without having to check error first.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
