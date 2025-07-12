package slices

func Map[T any, R any](col []T, fn func(elem T) R) []R {
	results := make([]R, len(col))
	for i, v := range col {
		results[i] = fn(v)
	}
	return results
}
