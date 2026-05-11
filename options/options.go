package options

import (
	"go.chrisrx.dev/x/structs"
)

type Option[T any] interface {
	Apply(T)
}

func New[T any](opts []Option[T]) T {
	v := structs.DefaultsFor[T]()
	for _, opt := range opts {
		opt.Apply(v)
	}
	return v
}
