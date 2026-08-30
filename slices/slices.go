//go:generate go tool aliaspkg -docs=all -ignore=Sort,Reverse,Max,Min

package slices

import (
	"cmp"
	"fmt"
	"hash/maphash"
	"slices"

	"go.chrisrx.dev/x/constraints"
	"go.chrisrx.dev/x/ptr"
)

// Filter returns the elements of col for which fn returns true.
func Filter[T any](col []T, fn func(elem T) bool) (result []T) {
	for _, v := range col {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterMap maps each element of col using fn and returns only the non-zero
// results.
func FilterMap[T any, R any](col []T, fn func(elem T) R) (result []R) {
	for _, v := range col {
		if v := fn(v); !ptr.IsZero(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterMap2 maps each element of col using fn and returns only elements where
// fn returns true.
func FilterMap2[T any, R any](col []T, fn func(elem T) (R, bool)) (result []R) {
	for _, v := range col {
		if v, ok := fn(v); ok {
			result = append(result, v)
		}
	}
	return result
}

// Find returns the first element in col for which fn returns true, or the zero
// value if none is found.
func Find[T any](col []T, fn func(elem T) bool) T {
	for _, v := range col {
		if fn(v) {
			return v
		}
	}
	return *new(T)
}

// FlatMap maps each element of col to a slice using fn and concatenates the
// results into a single slice.
func FlatMap[T any, R any](col []T, fn func(elem T) []R) []R {
	results := make([]R, 0)
	for _, elem := range col {
		results = append(results, fn(elem)...)
	}
	return results
}

// FlatMapErr maps each element of col to a slice using fn and concatenates the
// results into a single slice, stopping and returning the first error
// encountered.
func FlatMapErr[T any, R any](col []T, fn func(elem T) ([]R, error)) ([]R, error) {
	results := make([]R, 0)
	for _, elem := range col {
		elems, err := fn(elem)
		if err != nil {
			return nil, err
		}
		results = append(results, elems...)
	}
	return results, nil
}

// Map returns a new slice with each element of col transformed by fn.
func Map[T any, R any](col []T, fn func(elem T) R) []R {
	results := make([]R, len(col))
	for i, v := range col {
		results[i] = fn(v)
	}
	return results
}

// MapErr returns a new slice with each element of col transformed by fn,
// stopping and returning the first error encountered.
func MapErr[T any, R any](col []T, fn func(elem T) (R, error)) ([]R, error) {
	results := make([]R, len(col))
	for i, v := range col {
		var err error
		results[i], err = fn(v)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// MapEntries converts a slice to a map by applying fn to each element to
// produce a key-value pair.
func MapEntries[K comparable, V any, T any](col []T, fn func(elem T) (K, V)) map[K]V {
	result := make(map[K]V)
	for _, elem := range col {
		k, v := fn(elem)
		result[k] = v
	}
	return result
}

// MapEntriesErr converts a slice to a map by applying fn to each element to
// produce a key-value pair, stopping and returning the first error
// encountered.
func MapEntriesErr[K comparable, V any, T any](col []T, fn func(elem T) (K, V, error)) (map[K]V, error) {
	result := make(map[K]V)
	for _, elem := range col {
		k, v, err := fn(elem)
		if err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, nil
}

// Max returns the maximum value among vals, or the zero value if no arguments
// are provided.
func Max[T cmp.Ordered](vals ...T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	m := vals[0]
	for _, v := range vals[1:] {
		m = max(m, v)
	}
	return m
}

// Min returns the minimum value among vals, or the zero value if no arguments
// are provided.
func Min[T cmp.Ordered](vals ...T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	m := vals[0]
	for _, v := range vals[1:] {
		m = min(m, v)
	}
	return m
}

// N returns a slice of integers [0, n).
func N[T constraints.Integer](n T) []T {
	result := make([]T, n)
	for i := range int(n) {
		result[i] = T(i)
	}
	return result
}

// Partition splits s into two slices: left contains elements for which fn
// returns true, right contains the rest.
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

// Reverse reverses the elements of the slice in place. Unlike the standard
// library [slices.Reverse], this returns the reversed slice.
func Reverse[S ~[]E, E any](s S) S {
	slices.Reverse(s)
	return s
}

// Sort sorts a slice of any ordered type in ascending order. When sorting
// floating-point numbers, NaNs are ordered before other values.
func Sort[S ~[]E, E cmp.Ordered](x S) S {
	slices.Sort(x)
	return x
}

// Uniq returns only the unique elements of a slice.
func Uniq[S ~[]E, E any](s S) S {
	var h maphash.Hash
	hash := func(v any) uint64 {
		h.Reset()
		_, _ = fmt.Fprint(&h, v)
		return h.Sum64()
	}
	m := make(map[uint64]struct{})
	return slices.DeleteFunc(s, func(elem E) bool {
		n := hash(elem)
		if _, ok := m[n]; ok {
			return true
		}
		m[n] = struct{}{}
		return false
	})
}

// Or returns the first slice containing 1 or more elements.
func Or[S ~[]E, E any](slices ...S) S {
	var zero S
	for _, slice := range slices {
		if len(slice) > 0 {
			return slice
		}
	}
	return zero
}

// Truncate returns a new slice truncated to the provided upper index value.
func Truncate[S ~[]E, E any, N constraints.Integer](s S, upper N) S {
	return s[:min(len(s), int(upper))]
}

// Make returns s[i:j] with both bounds clamped to len(s), preventing
// out-of-range panics.
func Make[S ~[]E, E any, N constraints.Integer](s S, i, j N) S {
	return s[min(int(i), len(s)):min(int(j), len(s))]
}

func Of[S ~[]E, E any](elems ...E) S {
	return append(S{}, elems...)
}
