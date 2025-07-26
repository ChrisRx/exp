// Package set provides an implementation of a set which accepts any values,
// even if not comparable.
package set

import (
	"encoding/json"
	"iter"
	"maps"

	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/sync"
)

// Set holds a collection of unique values.
//
// It makes use of [maphash.Hash] to allow for adding values of any type, even
// ones that are uncomparable. The seed is set as a package global to ensure
// that all sets will produce hashes compatible for comparison.
//
// All methods lazily initialize the underlying storage, so the zero value of
// Set can be used.
type Set[T any] struct {
	hs hashset[T]

	once sync.Once
}

func New[T any](elems ...T) *Set[T] {
	s := &Set[T]{}
	s.init()
	for _, elem := range elems {
		s.hs.add(elem)
	}
	return s
}

func (set *Set[T]) init() {
	set.once.Do(func() {
		set.hs.reset()
	})
}

// Add adds an element to the set.
func (set *Set[T]) Add(elems ...T) {
	set.init()
	for _, elem := range elems {
		set.hs.add(elem)
	}
}

// Removes an element from the set, if present.
func (set *Set[T]) Remove(elems ...T) {
	set.init()
	for _, e := range elems {
		set.hs.remove(e)
	}
}

// All checks that the set contains all of the elements provided. If any are
// not part of the set, this returns false.
func (set *Set[T]) All(elems ...T) bool {
	set.init()
	return set.hs.all(elems...)
}

// Any checks if the set contains at least one of the provided elements.
func (set *Set[T]) Any(elems ...T) bool {
	set.init()
	return set.hs.any(elems...)
}

// Contains checks if the set contains at least one of the provided elements.
// This is an alias to [Set.Any].
func (set *Set[T]) Contains(elems ...T) bool {
	return set.Any(elems...)
}

// Equals compares all the elements of two sets for equality.
func (set *Set[T]) Equals(other *Set[T]) bool {
	return Compare(set, other)
}

// Len returns how many elements are current in the set.
func (set *Set[T]) Len() int {
	return len(set.hs.m)
}

// Size returns how many elements are current in the set. This is an alias to
// [Set.Len]
func (set *Set[T]) Size() int {
	return set.Len()
}

// IsEmpty checks if the set is empty.
func (set *Set[T]) IsEmpty() bool {
	return set.Len() == 0
}

// Clear removes all values from the set.
func (set *Set[T]) Clear() {
	set.once.Reset()
	set.init()
}

// Copy constructs a new set, copying the internal state of this set into the
// new returned set.
func (set *Set[T]) Copy() *Set[T] {
	var new Set[T]
	new.init()
	new.hs.m = maps.Clone(set.hs.m)
	return &new
}

// Each applies the provided function to every element of the set.
func (set *Set[T]) Each(fn func(T)) {
	set.init()
	set.hs.each(fn)
}

// List returns all the values of the set as a slice.
func (set *Set[T]) List() []T {
	return slices.Collect(set.Values())
}

// All returns all the values of the set as a sequence.
func (set *Set[T]) Values() iter.Seq[T] {
	set.init()
	return func(yield func(T) bool) {
		for _, elem := range set.hs.m {
			if !yield(elem) {
				return
			}
		}
	}
}

// Pop removes and returns one value from the set. The element returned is
// indeterministically selected.
func (set *Set[T]) Pop() T {
	set.init()
	return set.hs.pop()
}

// Difference returns a new set containing all of the elements in this set that
// are not present in the other set.
func (set *Set[T]) Difference(other *Set[T]) Set[T] {
	set.init()
	new := set.Copy()
	other.Each(func(elem T) {
		new.hs.remove(elem)
	})
	return ptr.From(new)
}

// Intersection returns a new set made up of all of the elements that are
// contained in both sets.
func (set *Set[T]) Intersection(other *Set[T]) (new Set[T]) {
	set.init()
	new.init()
	other.Each(func(elem T) {
		if set.hs.any(elem) {
			new.hs.add(elem)
		}
	})
	return
}

// Union returns a new set made up of all elements from each set.
func (set *Set[T]) Union(other *Set[T]) Set[T] {
	set.init()
	new := set.Copy()
	other.Each(func(elem T) {
		new.hs.add(elem)
	})
	return ptr.From(new)
}

func (set *Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.List())
}

func (set *Set[T]) UnmarshalJSON(data []byte) error {
	var values []T
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	set.Clear()
	set.Add(values...)
	return nil
}

// Compare compares two sets for equality, returning true if they contain all
// the same elements.
func Compare[T any](s1, s2 *Set[T]) bool {
	if len(s1.hs.m) != len(s2.hs.m) {
		return false
	}
	for k := range s1.hs.m {
		if _, ok := s2.hs.m[k]; !ok {
			return false
		}
	}
	return true
}
