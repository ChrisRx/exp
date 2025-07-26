package slices

import (
	"cmp"
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

func Clone[S ~[]E, E any](s S) S {
	return slices.Clone(s)
}

func Collect[E any](seq iter.Seq[E]) []E {
	return slices.Collect(seq)
}

func Compact[S ~[]E, E comparable](s S) S {
	return slices.Compact(s)
}

func CompactFunc[S ~[]E, E any](s S, eq func(E, E) bool) S {
	return slices.CompactFunc(s, eq)
}

func Concat[S ~[]E, E any](ss ...S) S {
	return slices.Concat(ss...)
}

func Contains[S ~[]E, E comparable](s S, v E) bool {
	return slices.Contains(s, v)
}

func ContainsFunc[S ~[]E, E any](s S, f func(E) bool) bool {
	return slices.ContainsFunc(s, f)
}

func Delete[S ~[]E, E any](s S, i, j int) S {
	return slices.Delete(s, i, j)
}

func DeleteFunc[S ~[]E, E any](s S, del func(E) bool) S {
	return slices.DeleteFunc(s, del)
}

func Grow[S ~[]E, E any](s S, n int) S {
	return slices.Grow(s, n)
}

func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	return slices.Insert(s, i, v...)
}

func Replace[S ~[]E, E any](s S, i, j int, v ...E) S {
	return slices.Replace(s, i, j, v...)
}

func Repeat[S ~[]E, E any](x S, count int) S {
	return slices.Repeat(x, count)
}

func Reverse[S ~[]E, E any](s S) S {
	slices.Reverse(s)
	return s
}

func Sort[S ~[]E, E cmp.Ordered](x S) S {
	slices.Sort(x)
	return x
}

func SortFunc[S ~[]E, E any](x S, cmp func(a, b E) int) {
	slices.SortFunc(x, cmp)
}

func Values[Slice ~[]E, E any](s Slice) iter.Seq[E] {
	return slices.Values(s)
}
