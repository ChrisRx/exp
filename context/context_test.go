package context_test

import (
	"fmt"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/context"
)

type ValueKey struct{}

type Value struct {
	s string
}

func TestKey(t *testing.T) {
	ctx := t.Context()

	t.Run("struct", func(t *testing.T) {
		k := context.Key[Value]()
		fmt.Printf("%s\n", k)
		assert.Equal(t, false, k.Has(ctx))

		ctx = k.WithValue(ctx, Value{s: "hi"})
		assert.Equal(t, true, k.Has(ctx))
		assert.Equal(t, "hi", k.Value(ctx).s)
	})

	t.Run("slice", func(t *testing.T) {
		k := context.Key[[]string]()
		ctx = k.WithValue(ctx, []string{"a", "b"})
		ctx = k.WithValue(ctx, append(k.Value(ctx), "c"))
		assert.Equal(t, []string{"a", "b", "c"}, k.Value(ctx))
	})

	t.Run("wrapped", func(t *testing.T) {
		k := context.Key[string]()

		ctx := t.Context()
		ctx = context.WithValue(ctx, k, "itworks")
		assert.Equal(t, "itworks", ctx.Value(k))
		assert.Equal(t, "itworks", k.Value(ctx))
	})
}
