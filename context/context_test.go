package context_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.chrisrx.dev/x/context"
)

type ValueKey struct{}

type Value struct {
	s string
}

func TestFrom(t *testing.T) {
	ctx := context.WithValue(t.Context(), ValueKey{}, Value{s: "some value"})
	v := context.From[ValueKey, Value](ctx)
	assert.Equal(t, "some value", v.s)
}

func TestKey(t *testing.T) {
	ctx := t.Context()

	k := context.Key[Value]()
	fmt.Printf("%s\n", k)
	assert.Equal(t, false, k.Has(ctx))

	ctx = k.WithValue(ctx, Value{s: "hi"})
	assert.Equal(t, true, k.Has(ctx))
	assert.Equal(t, "hi", k.Value(ctx).s)
}
