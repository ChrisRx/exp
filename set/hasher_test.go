package set

import (
	"fmt"
	"hash/maphash"
	"testing"
)

var seed = maphash.MakeSeed()

func hash(v any) uint64 {
	h := new(maphash.Hash)
	h.SetSeed(seed)
	_, _ = fmt.Fprint(h, v)
	return h.Sum64()
}

type unsafeHasher struct {
	h maphash.Hash
}

func (h *unsafeHasher) Hash1(v any) uint64 {
	h.h.Reset()
	_, _ = fmt.Fprint(&h.h, v)
	return h.h.Sum64()
}

func (h unsafeHasher) Hash2(v any) uint64 {
	h.h.Reset()
	_, _ = fmt.Fprint(&h.h, v)
	return h.h.Sum64()
}

func BenchmarkHasher(b *testing.B) {
	b.Run("safe reset", func(b *testing.B) {
		h := newSafeHasher(maphash.MakeSeed())
		b.ResetTimer()
		var i int
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				h.Hash(i)
				i++
			}
		})
	})

	b.Run("safe new", func(b *testing.B) {
		var i int
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hash(i)
				i++
			}
		})
	})

	b.Run("unsafe pointer receiver", func(b *testing.B) {
		h := &unsafeHasher{}
		h.h.SetSeed(maphash.MakeSeed())
		b.ResetTimer()
		for i := range b.N {
			h.Hash1(i)
		}
	})

	b.Run("unsafe value receiver", func(b *testing.B) {
		h := unsafeHasher{}
		h.h.SetSeed(maphash.MakeSeed())
		b.ResetTimer()
		for i := range b.N {
			h.Hash2(i)
		}
	})

	b.Run("unsafe value receiver2", func(b *testing.B) {
		h := &unsafeHasher{}
		h.h.SetSeed(maphash.MakeSeed())
		b.ResetTimer()
		for i := range b.N {
			h.Hash2(i)
		}
	})
}
