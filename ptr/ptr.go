// Package ptr contains functions for handling pointer values.
package ptr

import (
	"reflect"
)

// To returns a pointer value for the provided value. This is useful for
// situations where the address cannot be taken of something directly.
// e.g.	To("string literal").
func To[T any](v T) *T {
	return &v
}

// From returns the underlying value of the provided pointer value. If a nil
// value is provided, the zero value of the type will be returned.
func From[T any](v *T) T {
	if v != nil {
		return *v
	}
	return *new(T)
}

// Nullable returns the provided value unless it is the zero value for the
// type. This is useful when composing structs where an empty value should be
// assigned as a nil.
func Nullable[T *E, E any](v T) T {
	if IsZero(v) {
		return nil
	}
	return v
}

// ToNullable returns a pointer to the provided value. If the value is the zero
// value, a nil is returned.
func ToNullable[T any](v T) *T {
	if IsZero(v) {
		return nil
	}
	return To(v)
}

// NullableSlice returns the provided slice unless it is empty, which returns
// nil.
func NullableSlice[S []E, E any](v S) S {
	if len(v) == 0 {
		return nil
	}
	return v
}

// Equal returns true if the underlying value of the provided pointer values
// are equal. This is intended to determine equality of the concrete types so
// zero value and a nil pointer will always return as equal.
func Equal[T comparable](a, b *T) bool {
	return From(a) == From(b)
}

// IsZero returns true if value is the zero value for the type.
func IsZero(v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return reflect.Indirect(rv).IsZero()
}
