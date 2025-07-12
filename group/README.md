[![Go Reference](https://pkg.go.dev/badge/go.chrisrx.dev/group.svg)](https://pkg.go.dev/go.chrisrx.dev/group)

# group

group is a library for managing pools of goroutines. It has been adapted from [errgoup](https://pkg.go.dev/golang.org/x/sync/errgroup) with small improvements to the API.

## Usage


## Simple

This will create a new group and start 10 goroutines:

```go
g := group.New(ctx)
for i := range 10 {
	g.Go(func(ctx context.Context) error {
		fmt.Printf("loop %d\n", i)
		return nil
	})
}
if err := g.Wait(); err != nil {
	log.Fatal(err)
}
```

The parent context provided is used to create a child context that group uses internally, which is passed through to each goroutine. If any goroutine produces an error, this child context is canceled, allowing the other goroutines to stop/cleanup:

## Bounded concurrency

The option `WithLimit` can be passed to the group constructor to establish a bound on concurrency:


```go
g := group.New(ctx, group.WithLimit(2))
for i := range 10 {
	g.Go(func(ctx context.Context) error {
		fmt.Printf("loop %d\n", i)
		return nil
	})
}
if err := g.Wait(); err != nil {
	log.Fatal(err)
}
```

Here, only 2 goroutines will ever be running at a given time.


## Method chaining

A group can also be setup using method chaining:

```go
if err := group.New(ctx).Go(func(ctx context.Context) error {
	fmt.Printf("goroutine 1\n")
	time.Sleep(1 * time.Second)
	return nil
}).Go(func(ctx context.Context) error {
	fmt.Printf("goroutine 2\n")
	time.Sleep(5 * time.Second)
	return nil
}).Wait(); err != nil {
	log.Fatal(err)
}
```

## Results


```go
g := group.NewResultGroup[int](ctx)
for i := range 10 {
	g.Go(func(ctx context.Context) (int, error) {
		return i, nil
	})
}
for v, err := range g.Get() {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)
}
```

## Getting a future result

```go
g := group.NewResultGroup[string](ctx)
result := g.Go(func(ctx context.Context) (string, error) {
	time.Sleep(500 * time.Millisecond)
	return "future value", nil
})
v, err := result.Get()
if err != nil {
	log.Fatal(err)
}
if err := g.Wait(); err != nil {
	log.Fatal(err)
}
```

## Notes

* https://github.com/golang/go/issues/57534
