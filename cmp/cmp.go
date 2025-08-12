//go:generate go tool aliaspkg -docs=all

package cmp

import (
	"slices"
)

func All[T comparable](S ...T) bool {
	var zero T
	return !slices.Contains(S, zero)
}

func Any[T comparable](S ...T) bool {
	var zero T
	for _, elem := range S {
		if elem != zero {
			return true
		}
	}
	return false
}
