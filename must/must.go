// Package must provides functions for handling errors and panics.
package must

// Ok takes a function that returns a value and error and returns only a
// value, otherwise it panics.
func Ok[T any](v T, err error) (zero T) {
	if err != nil {
		panic(err)
	}
	return v
}

func Get0[T1, T2 any](v1 T1, v2 T2) T1 { return v1 }
func Get1[T1, T2 any](v1 T1, v2 T2) T2 { return v2 }
