//go:generate go tool aliaspkg -docs=all

package maps

func ToSlice[K comparable, V any, R any](m map[K]V, fn func(K, V) R) []R {
	result := make([]R, len(m))
	var i int
	for k, v := range m {
		result[i] = fn(k, v)
		i++
	}
	return result
}

func Filter[K comparable, V any](m map[K]V, fn func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if fn(k, v) {
			result[k] = v
		}
	}
	return result
}

func Map[K1, K2 comparable, V1, V2 any](m map[K1]V1, fn func(K1, V1) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2)
	for k, v := range m {
		k, v := fn(k, v)
		result[k] = v
	}
	return result
}
