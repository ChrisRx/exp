package sort

import (
	"iter"
	"sort"

	"go.chrisrx.dev/x/cmp"
)

type Entry[K comparable, V cmp.Ordered] struct {
	Key   K
	Value V
}

type SortMap[K comparable, V cmp.Ordered] []Entry[K, V]

func (m SortMap[K, V]) Len() int           { return len(m) }
func (m SortMap[K, V]) Less(i, j int) bool { return m[i].Value < m[j].Value }
func (m SortMap[K, V]) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func (m SortMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, kv := range m {
			if !yield(kv.Key, kv.Value) {
				return
			}
		}
	}
}

func (m SortMap[K, V]) Limit(n int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var i int
		for k, v := range m.All() {
			i++
			if i > n {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

func Map[K comparable, V cmp.Ordered](m map[K]V) SortMap[K, V] {
	sm := make(SortMap[K, V], len(m))
	i := 0
	for k, v := range m {
		sm[i] = Entry[K, V]{k, v}
		i++
	}
	sort.Sort(sm)
	return sm
}
