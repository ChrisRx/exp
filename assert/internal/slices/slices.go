package slices

import (
	"cmp"
	"slices"

	"go.chrisrx.dev/x/constraints"
)

// assert is going to be used in pretty much every package so it needs to copy
// code that might exist in those packages already to prevent an import cycle.

func Map[T any, R any](col []T, fn func(elem T) R) []R {
	results := make([]R, len(col))
	for i, v := range col {
		results[i] = fn(v)
	}
	return results
}

func Filter[T any](col []T, fn func(elem T) bool) []T {
	results := make([]T, 0)
	for _, v := range col {
		if fn(v) {
			results = append(results, v)
		}
	}
	return results
}

func N[T constraints.Integer](n T) []T {
	result := make([]T, n)
	for i := range int(n) {
		result[i] = T(i)
	}
	return result
}

func Contains[S ~[]E, E comparable](s S, v E) bool {
	return slices.Contains(s, v)
}

func Concat[S ~[]E, E any](s ...S) S {
	return slices.Concat(s...)
}

func Max[S ~[]E, E cmp.Ordered](x S) E {
	return slices.Max(x)
}
