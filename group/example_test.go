package group_test

import (
	"context"
	"fmt"
	"log"

	"go.chrisrx.dev/x/group"
)

func ExampleGroup() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := group.New(ctx)
	for range 10 {
		g.Go(func(ctx context.Context) error {
			fmt.Printf("starting goroutine ...\n")
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("done\n")

	// Output: starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// starting goroutine ...
	// done
}

func ExampleResultGroup() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := group.NewResultGroup[string](ctx)
	for range 5 {
		g.Go(func(ctx context.Context) (string, error) {
			return "result", nil
		})
	}
	for v, err := range g.Get() {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(v)
	}

	fmt.Printf("done\n")

	// Output: result
	// result
	// result
	// result
	// result
	// done
}
