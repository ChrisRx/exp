package context

import (
	"context"
	"fmt"

	"go.chrisrx.dev/x/safe"
)

//go:generate go tool aliaspkg -docs=all

type key[V any] struct {
	_ safe.NoCopy
}

// Key creates a new key used for loading and storing values in a
// [context.Context]. The key itself is a pointer to the underlying key type,
// therefore should not be used directly with functions like
// [context.Context.Value]. Instead, the methods for key should be used to
// load/store values.
//
// Key is intended to make package level context keys that can be shared by
// other packages.
func Key[V any]() *key[V] {
	return &key[V]{}
}

// Has checks the provided [context.Context] if this key has a value set.
func (k *key[V]) Has(ctx context.Context) bool {
	_, ok := ctx.Value(k).(V)
	return ok
}

// Value returns the value stored in the provided [context.Context]. If no
// value is set, the zero value for the parameterized type for this key is
// returned.
func (k *key[V]) Value(ctx context.Context) V {
	if v, ok := ctx.Value(k).(V); ok {
		return v
	}
	var zero V
	return zero
}

// WithValue returns a new [context.Context] derived from the provided
// [context.Context] and value.
func (k *key[V]) WithValue(parent context.Context, value V) context.Context {
	return context.WithValue(parent, k, value)
}

func (k *key[V]) String() string {
	return fmt.Sprintf("context.Key[%T]", k)
}
