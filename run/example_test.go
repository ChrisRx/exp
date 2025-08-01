package run_test

import (
	"context"
	"fmt"
	"time"

	"go.chrisrx.dev/x/run"
)

func ExampleEvery() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	run.Every(ctx, func() {
		fmt.Println("doing some work")
	}, 450*time.Millisecond)

	// Output: doing some work
	// doing some work
	// doing some work
	// doing some work
}
