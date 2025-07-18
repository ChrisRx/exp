package slices

import (
	"iter"
	"slices"
)

func All[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] {
	return slices.All(s)
}

func Backward[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] {
	return slices.Backward(s)
}

func Chunk[Slice ~[]E, E any](s Slice, n int) iter.Seq[Slice] {
	return slices.Chunk(s, n)
}

func Collect[E any](seq iter.Seq[E]) []E {
	return slices.Collect(seq)
}

func Values[Slice ~[]E, E any](s Slice) iter.Seq[E] {
	return slices.Values(s)
}

func Contains[S ~[]E, E comparable](s S, v E) bool {
	return slices.Contains(s, v)
}

func ContainsFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	return slices.ContainsFunc(s, f)
}

func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	return slices.Insert(s, i, v...)
}

func Delete[S ~[]E, E any](s S, i, j int) S {
	return slices.Delete(s, i, j)
}

func DeleteFunc[S ~[]E, E any](s S, del func(E) bool) S {
	return slices.DeleteFunc(s, del)
}

func Replace[S ~[]E, E any](s S, i, j int, v ...E) S {
	return slices.Replace(s, i, j, v...)
}

func Clone[S ~[]E, E any](s S) S {
	return slices.Clone(s)
}

func Compact[S ~[]E, E comparable](s S) S {
	return slices.Compact(s)
}

func CompactFunc[S ~[]E, E any](s S, eq func(E, E) bool) S {
	return slices.CompactFunc(s, eq)
}

func Grow[S ~[]E, E any](s S, n int) S {
	return slices.Grow(s, n)
}

func Reverse[S ~[]E, E any](s S) S {
	slices.Reverse(s)
	return s
}

func Concat[S ~[]E, E any](ss ...S) S {
	return slices.Concat(ss...)
}

func Repeat[S ~[]E, E any](x S, count int) S {
	return slices.Repeat(x, count)
}
