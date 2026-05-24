# context

Package `context` is a drop-in replacement for the standard library [`context`](https://pkg.go.dev/context) package. Swapping the import path gives you all standard library functions plus type-safe generic keys and signal-based graceful shutdown — no other code changes required.

## Type-safe context keys

`Key[V]()` returns a typed key that stores and retrieves values from a `context.Context` without requiring unexported struct types or manual type assertions.

```go
import "go.chrisrx.dev/x/context"

// Declare a package-level key.
var AuthKey = context.Key[*Auth]()

// Store a value.
ctx = AuthKey.WithValue(ctx, &Auth{Token: "Bearer ..."})

// Retrieve a value — returns the zero value if not set.
auth := AuthKey.Value(ctx)

// Check presence without retrieving.
if AuthKey.Has(ctx) { ... }

// Call a function only when the value is present.
AuthKey.ValueFunc(ctx, func(a *Auth) {
    log.Println("auth:", a.Token)
})
```

The key itself is a pointer to an unexported generic type, so two `Key[V]()` calls always produce distinct keys even for the same type `V`.

## Graceful shutdown

`Shutdown` returns a `ShutdownContext` that listens for OS signals (`SIGINT`, `SIGTERM`) and executes registered handlers in FIFO order — one per signal received. When all handlers have run, or any handler completes, the context is cancelled and default signal handling is restored.

```go
ctx := context.Shutdown()

// First SIGINT: attempt a graceful drain.
ctx.AddHandler(func() {
    fmt.Println("shutting down gracefully...")
    server.Shutdown(context.Background())
})

// Second SIGINT: force exit.
ctx.AddHandler(func() {
    fmt.Println("forcing exit")
    os.Exit(1)
})

// Block until the context is done.
ctx.Wait()
```

Cleanup functions registered with `AddCleanup` run when the context is garbage collected. Unlike handlers they are not triggered by signals — they are a last-resort cleanup path with no ordering guarantees relative to program exit.

```go
ctx.AddCleanup(func() {
    db.Close()
})
```

`AddHandler` and `AddCleanup` are also available as package-level functions that accept any `context.Context`, logging a warning when the context is not a `ShutdownContext`.
