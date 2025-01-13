package xiter

import "iter"

func Chunk[V any](seq iter.Seq2[V, error], n int64) iter.Seq2[[]V, error] {
	return func(yield func([]V, error) bool) {
		if n <= 0 {
			return
		}
		chunk := make([]V, 0, n)
		for v, err := range seq {
			if err != nil {
				yield(nil, err)
				return
			}

			chunk = append(chunk, v)
			if len(chunk) >= int(n) {
				if !yield(chunk, nil) {
					return
				}
				chunk = make([]V, 0, n)
				continue
			}
		}
		if len(chunk) > 0 {
			if !yield(chunk, nil) {
				return
			}
		}
	}
}
