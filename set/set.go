package set

import (
	"iter"

	"go.chrisrx.dev/x/sync"
)

type Set[K comparable] struct {
	m    map[K]struct{}
	once sync.OnceAgain
}

func (s *Set[K]) Add(v K) {
	s.init()
	s.m[v] = struct{}{}
}

func (s *Set[K]) Clear() {
	s.once.Reset()
}

func (s *Set[K]) Contains(v K) bool {
	s.init()
	_, ok := s.m[v]
	return ok
}

func (s *Set[K]) Len() int {
	return len(s.m)
}

func (s *Set[K]) List() iter.Seq[K] {
	s.init()
	return func(yield func(K) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}

func (s *Set[K]) Remove(v K) {
	s.init()
	delete(s.m, v)
}

func (s *Set[K]) init() {
	s.once.Do(func() {
		s.m = make(map[K]struct{})
	})
}
