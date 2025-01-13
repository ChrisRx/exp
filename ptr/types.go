package ptr

import (
	"golang.org/x/exp/constraints"
)

// Int returns a int pointer value for the provided integer value.
func Int[T constraints.Integer](v T) *int {
	return To(int(v))
}

// Int32 returns an int32 pointer value for the provided integer value.
func Int32[T constraints.Integer](v T) *int32 {
	return To(int32(v))
}

// Int64 returns an int64 pointer value for the provided integer value.
func Int64[T constraints.Integer](v T) *int64 {
	return To(int64(v))
}

// Uint returns a uint pointer value for the provided integer value.
func Uint[T constraints.Integer](v T) *uint {
	return To(uint(v))
}

// Uint32 returns an uint32 pointer value for the provided integer value.
func Uint32[T constraints.Integer](v T) *uint32 {
	return To(uint32(v))
}

// Uint64 returns an uint64 pointer value for the provided integer value.
func Uint64[T constraints.Integer](v T) *uint64 {
	return To(uint64(v))
}

// String returns a string pointer value for the provided string value.
func String(s string) *string {
	return To(s)
}

// Bool returns a bool pointer value for the provided bool value.
func Bool(b bool) *bool {
	return To(b)
}
