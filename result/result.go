package result

import "iter"

type Of[T any] struct {
	v   T
	err error
}

func Ok[T any](v T) Of[T] {
	return Of[T]{v: v}
}

func Err[T any](err error) Of[T] {
	return Of[T]{err: err}
}

func (r Of[T]) Get() (T, error) {
	return r.v, r.err
}

func (r Of[T]) MustGet() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.v
}

func (r Of[T]) Err() error {
	return r.err
}

func Unwrap[T any](seq iter.Seq[Of[T]]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for result := range seq {
			v, err := result.Get()
			if err != nil {
				yield(*new(T), err)
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}
