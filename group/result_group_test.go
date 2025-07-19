package group_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/future"
	"go.chrisrx.dev/x/group"
)

func TestResultGroup(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		g := group.NewResultGroup[string](t.Context())
		results := make([]future.Value[string], 0)
		for i := range 10 {
			results = append(results, g.Go(func(ctx context.Context) (string, error) {
				time.Sleep(500 * time.Millisecond)
				return fmt.Sprintf("loop %d", i), nil
			}))
		}

		time.Sleep(600 * time.Millisecond)

		for _, result := range results {
			v, err := result.Get()
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(v)
		}
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("iterator", func(t *testing.T) {
		g := group.NewResultGroup[string](t.Context())
		for i := range 10 {
			g.Go(func(ctx context.Context) (string, error) {
				n := rand.IntN(300-100+1) + 100
				time.Sleep(time.Duration(n) * time.Millisecond)
				return fmt.Sprintf("loop %d", i), nil
			})
		}

		var i int
		for v, err := range g.Get() {
			if err != nil {
				t.Fatal(err)
			}
			i++
			fmt.Println(v)
		}
		assert.Equal(t, 10, i)
	})

	t.Run("iterator with limit", func(t *testing.T) {
		g := group.NewResultGroup[string](t.Context(), group.WithLimit(8))
		for i := range 10 {
			g.Go(func(ctx context.Context) (string, error) {
				n := rand.IntN(300-100+1) + 100
				time.Sleep(time.Duration(n) * time.Millisecond)
				return fmt.Sprintf("loop %d", i), nil
			})
		}

		var i int
		for v, err := range g.Get() {
			if err != nil {
				t.Fatal(err)
			}
			i++
			fmt.Println(v)
		}
		assert.Equal(t, 10, i)
	})
}
