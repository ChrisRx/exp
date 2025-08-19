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
