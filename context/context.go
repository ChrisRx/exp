package context

import (
	"context"
	"fmt"
)

type Context = context.Context

func WithValue(parent context.Context, key, val any) context.Context {
	return context.WithValue(parent, key, val)
}

type key[V any] struct{}

func Key[V any]() key[V] {
	return key[V]{}
}

func (k key[V]) String() string {
	return fmt.Sprintf("context.Key[%T]", k)
}

func (k key[V]) Has(ctx context.Context) bool {
	_, ok := ctx.Value(k).(V)
	return ok
}

func (k key[V]) Value(ctx context.Context) V {
	if v, ok := ctx.Value(k).(V); ok {
		return v
	}
	var zero V
	return zero
}

func (k key[V]) WithValue(parent context.Context, value V) context.Context {
	return context.WithValue(parent, k, value)
}

func From[K comparable, V any](ctx context.Context) V {
	if v, ok := ctx.Value(*new(K)).(V); ok {
		return v
	}
	var zero V
	return zero
}
