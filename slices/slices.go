package slices

import "go.chrisrx.dev/x/constraints"

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
