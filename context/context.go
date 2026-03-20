package context

import (
	"context"
	"fmt"
	"os"

	"go.chrisrx.dev/x/log/slog"
	"go.chrisrx.dev/x/safe"
)

//go:generate go tool aliaspkg -docs=all

// ContextKey represents a key that stores values in a [context.Context].
type ContextKey[V any] interface {
	Has(ctx context.Context) bool
	Value(ctx context.Context) V
	ValueFunc(ctx context.Context, fn func(V)) bool
	WithValue(parent context.Context, value V) context.Context
}

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
func Key[V any]() ContextKey[V] {
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

// ValueFunc calls a function with the value stored in the provided
// [context.Context]. It returns true when the value is set. If no value is
// set, the function will not be called and this will return false.
func (k *key[V]) ValueFunc(ctx context.Context, fn func(V)) bool {
	v, ok := ctx.Value(k).(V)
	if ok {
		fn(v)
		return true
	}
	logger.Debug("context does not contain value", slog.Stringf("type", "%T", *new(V)))
	return false
}

// WithValue returns a new [context.Context] derived from the provided
// [context.Context] and value.
func (k *key[V]) WithValue(parent context.Context, value V) context.Context {
	return context.WithValue(parent, k, value)
}

func (k *key[V]) String() string {
	return fmt.Sprintf("context.Key[%T]", k)
}

var (
	// internal debug logger
	lvl    = new(slog.LevelVar)
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
)
