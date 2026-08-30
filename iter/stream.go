package iter

import (
	"fmt"
	"hash/maphash"

	"go.chrisrx.dev/x/slices"
)

// Stream is a [Seq] that supports chaining lazy operations, such as
// [Stream.Map] and [Stream.Filter], followed by a terminal operation that
// consumes the sequence, such as [Stream.Collect] or [Stream.Reduce].
// Chained operations are not applied until a terminal operation runs.
type Stream[V any] Seq[V]

// From converts a [Seq] into a [Stream].
func From[V any](seq Seq[V]) Stream[V] {
	return Stream[V](seq)
}

// Of returns a [Stream] over the elements of elems.
func Of[V any](elems []V) Stream[V] {
	return From(slices.Values(elems))
}

// Map returns a [Stream] over seq with fn applied to each value.
func Map[V, R any](seq Seq[V], fn func(V) R) Stream[R] {
	return From(seq).Map[R](fn)
}

// FlatMap returns a [Stream] over seq with fn applied to each value and the
// resulting slices flattened into a single sequence.
func FlatMap[V, R any](seq Seq[V], fn func(V) []R) Stream[R] {
	return From(seq).FlatMap[R](fn)
}

// Filter returns a [Stream] over the values in seq for which fn returns
// true.
func Filter[V any](seq Seq[V], fn func(V) bool) Stream[V] {
	return From(seq).Filter(fn)
}

// Filter returns a [Stream] over the values of s for which fn returns true.
func (s Stream[V]) Filter(fn func(V) bool) Stream[V] {
	return func(yield func(V) bool) {
		for v := range s {
			if !fn(v) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Map returns a [Stream] with fn applied to each value of s.
func (s Stream[V]) Map[R any](fn func(V) R) Stream[R] {
	return func(yield func(R) bool) {
		for v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// FlatMap returns a [Stream] with fn applied to each value of s and the
// resulting slices flattened into a single sequence.
func (s Stream[V]) FlatMap[R any](fn func(V) []R) Stream[R] {
	return func(yield func(R) bool) {
		for v := range s {
			for _, r := range fn(v) {
				if !yield(r) {
					return
				}
			}
		}
	}
}

// Skip returns a [Stream] over s with the first n values omitted.
func (s Stream[V]) Skip(n int) Stream[V] {
	return func(yield func(V) bool) {
		for i, v := range Enumerate(s.Values()) {
			if i >= n {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Uniq returns a [Stream] over s with duplicate values removed, preserving
// the order in which each distinct value first appears. Values are
// compared using a hash of their formatted ([fmt.Fprint]) representation,
// so V need not be comparable.
func (s Stream[V]) Uniq() Stream[V] {
	var h maphash.Hash
	hash := func(v any) uint64 {
		h.Reset()
		_, _ = fmt.Fprint(&h, v)
		return h.Sum64()
	}
	m := make(map[uint64]struct{})
	return func(yield func(V) bool) {
		for v := range s {
			n := hash(v)
			if _, ok := m[n]; ok {
				continue
			}
			if !yield(v) {
				return
			}
			m[n] = struct{}{}
		}
	}
}

// Find returns the first value of s for which fn returns true, or the zero
// value of V if no value satisfies fn. Find is a terminal operation that
// stops consuming s as soon as a match is found.
func (s Stream[V]) Find(fn func(V) bool) V {
	for v := range s {
		if fn(v) {
			return v
		}
	}
	return *new(V)
}

// Fold reduces s to a single value of type A by repeatedly applying fn,
// starting from the zero value of A as the initial accumulator.
func (s Stream[V]) Fold[A any](fn func(A, V) A) A {
	var acc A
	for v := range s {
		acc = fn(acc, v)
	}
	return acc
}

// Reduce reduces s to a single value of type V by repeatedly applying fn,
// starting from the zero value of V as the initial accumulator.
func (s Stream[V]) Reduce(fn func(V, V) V) V {
	return s.Fold[V](fn)
}

// Next returns the first value of s, or the zero value of V if s is empty.
func (s Stream[V]) Next() V {
	for v := range s {
		return v
	}
	return *new(V)
}

// Last returns the last value of s, or the zero value of V if s is empty.
// It consumes s in its entirety.
func (s Stream[V]) Last() V {
	var last V
	for v := range s {
		last = v
	}
	return last
}

// Take collects every value of s for which fn returns true into a slice.
func (s Stream[V]) Take(fn func(V) bool) (results []V) {
	for v := range s {
		if fn(v) {
			results = append(results, v)
		}
	}
	return
}

// All reports whether fn returns true for every value of s. It reports
// true for an empty stream and stops consuming s as soon as fn returns
// false for a value.
func (s Stream[V]) All(fn func(V) bool) bool {
	for v := range s {
		if !fn(v) {
			return false
		}
	}
	return true
}

// Any reports whether fn returns true for at least one value of s. It
// stops consuming s as soon as fn returns true for a value.
func (s Stream[V]) Any(fn func(V) bool) bool {
	for v := range s {
		if fn(v) {
			return true
		}
	}
	return false
}

// Each calls fn for every value of s.
func (s Stream[V]) Each(fn func(V)) {
	for v := range s {
		fn(v)
	}
}

// Values returns s as a plain [Seq].
func (s Stream[V]) Values() Seq[V] {
	return Seq[V](s)
}

// Collect gathers every value of s into a slice.
func (s Stream[V]) Collect() []V {
	return slices.Collect(Seq[V](s))
}

// SortFunc collects s into a slice and sorts it in place using fn to
// compare elements, in the same manner as [slices.SortFunc].
func (s Stream[V]) SortFunc(fn func(a, b V) int) []V {
	items := s.Collect()
	slices.SortFunc(items, fn)
	return items
}
