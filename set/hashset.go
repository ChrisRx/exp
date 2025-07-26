package set

import (
	"go.chrisrx.dev/x/slices"
)

// hashset is an internal implementation of a set using maphash.
type hashset[T any] struct {
	m map[uint64]T
}

// reset initializes the internal map storage and creates a new hasher with the
// package global seed value.
func (hs *hashset[T]) reset() {
	hs.m = make(map[uint64]T)
}

func (hs *hashset[T]) add(elem T) {
	h := hasher.Hash(elem)
	if _, ok := hs.m[h]; !ok {
		hs.m[h] = elem
	}
}

func (hs *hashset[T]) remove(elem T) {
	delete(hs.m, hasher.Hash(elem))
}

func (hs *hashset[T]) all(elems ...T) bool {
	for _, elem := range elems {
		if !hs.any(elem) {
			return false
		}
	}
	return true
}

func (hs *hashset[T]) any(elems ...T) bool {
	return slices.ContainsFunc(elems, func(elem T) bool {
		_, ok := hs.m[hasher.Hash(elem)]
		return ok
	})
}

func (hs *hashset[T]) each(fn func(T)) {
	for _, elem := range hs.m {
		fn(elem)
	}
}

func (hs *hashset[T]) pop() T {
	for _, elem := range hs.m {
		defer hs.remove(elem)
		return elem
	}
	var zero T
	return zero
}
