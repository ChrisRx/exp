package iter

import (
	"go.chrisrx.dev/x/constraints"
)

//go:generate go tool aliaspkg -docs=all

// N returns a [Stream] over the integers [0, n).
func N[T constraints.Integer](n T) Stream[T] {
	return func(yield func(T) bool) {
		for i := range int(n) {
			if !yield(T(i)) {
				return
			}
		}
	}
}

// Enumerate returns an iterator over seq that pairs each value with its
// index, starting at 0.
func Enumerate[T any](seq Seq[T]) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var i int
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}

// Find returns the first value in seq for which fn returns true, or the
// zero value of T if no value satisfies fn.
func Find[T any](seq Seq[T], fn func(T) bool) T {
	return From(seq).Find(fn)
}
