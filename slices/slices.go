package slices

import (
	"cmp"
	"slices"

	"go.chrisrx.dev/x/constraints"
)

//go:generate go tool aliaspkg -ignore Sort,Reverse

func Map[T any, R any](col []T, fn func(elem T) R) []R {
	results := make([]R, len(col))
	for i, v := range col {
		results[i] = fn(v)
	}
	return results
}

func Filter[T any](col []T, fn func(elem T) bool) (result []T) {
	for _, v := range col {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func Partition[S ~[]E, E any](s S, fn func(E) bool) (left, right S) {
	for _, elem := range s {
		if fn(elem) {
			left = append(left, elem)
		} else {
			right = append(right, elem)
		}
	}
	return
}

func N[T constraints.Integer](n T) []T {
	result := make([]T, n)
	for i := range int(n) {
		result[i] = T(i)
	}
	return result
}

func FlatMap[T any, R any](col []T, fn func(elem T) []R) []R {
	results := make([]R, 0)
	for _, elem := range col {
		results = append(results, fn(elem)...)
	}
	return results
}

// Sort sorts a slice of any ordered type in ascending order.
// When sorting floating-point numbers, NaNs are ordered before other values.
func Sort[S ~[]E, E cmp.Ordered](x S) S {
	slices.Sort(x)
	return x
}

// Reverse reverses the elements of the slice in place.
func Reverse[S ~[]E, E any](s S) S {
	slices.Reverse(s)
	return s
}
