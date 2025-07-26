package set

import (
	"fmt"
	"hash/maphash"
	"sync"
)

// This is the instance of safeHasher used within the set package. The seed
// used for set hash functions is global to ensure that comparisons are stable
// across different sets.
var hasher = newSafeHasher(maphash.MakeSeed())

// safeHasher is a thread-safe implementation of [maphash.Hash] used to create
// hashes of arbitrary byte sequences.
type safeHasher struct {
	mu sync.Mutex
	h  maphash.Hash
}

func newSafeHasher(seed maphash.Seed) *safeHasher {
	var h safeHasher
	h.mu.Lock()
	defer h.mu.Unlock()
	h.h.SetSeed(seed)
	return &h
}

func (h *safeHasher) Hash(v any) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.h.Reset()
	// Calls to (*maphash.Hash).Write do not produce an error, so checking the
	// error here is not necessary.
	_, _ = fmt.Fprint(&h.h, v)
	return h.h.Sum64()
}
