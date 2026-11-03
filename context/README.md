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

`Shutdown` returns a `ShutdownContext` that listens for OS signals (`SIGINT`, `SIGTERM`). It's two contexts:

- The `ShutdownContext` itself is the work context, the one you pass around. The first signal cancels it, so work notices and winds down, gracefully if it likes.
- `Program()` is its parent, the program's own context. It's canceled last, by `Close`, after the cleanup functions run. Things a graceful shutdown still needs, like a child process, belong to it.

Each further signal runs the next handler added with `AddHandler`, in FIFO order, to escalate when a graceful shutdown takes too long. Once there are none left, the next signal restores default signal handling and is raised again, ending the process. With no handlers, the first signal lets a program that waits on its context exit normally, and a second ends it.

```go
ctx := context.Shutdown()
defer ctx.Close()

// The server's process belongs to the program, so canceling the work
// context doesn't kill it while connections are draining.
server := StartServer(ctx.Program())

// Second signal: stop without waiting for the connections.
ctx.AddHandler(func() {
    server.HardStop()
})

// Serves until the first signal, then drains the connections.
server.Run(ctx)
```

See `ExampleShutdown_childProcess` for this running end to end.

The work context is canceled with a `SignalError` as its cause, so code that only has a `context.Context` can tell why it stopped. `SignalCause` returns the signal, and `AfterSignal` is like `AfterFunc` but only runs when a signal canceled the context, so a normal `Close` doesn't trigger it:

```go
context.AfterSignal(ctx, func(sig os.Signal) {
    switch sig {
    case os.Interrupt:
        fmt.Printf("\r%v received, shutting down gracefully, press again to stop now ...\n", sig)
    default:
        fmt.Printf("received %v, shutting down gracefully ...\n", sig)
    }
})
```

To guard against an accidental CTRL+C, `Intercept` adds a function that runs on an interrupt in place of canceling the work context. Each intercept consumes one interrupt, in FIFO order, and only before the work context is canceled. `InterceptTimeout` restores the intercept when no other interrupt follows within the timeout. Only interrupts are intercepted: a SIGTERM, as sent by Kubernetes, systemd or `docker stop`, comes once before a SIGKILL, so it always goes straight to canceling.

```go
ctx.InterceptTimeout(func() {
    fmt.Println("\rCTRL+C received, press again to shut down gracefully ...")
}, 5*time.Second)
```

The ladder is then: intercepts, canceling the work context, handlers, and re-raising the signal. See `ExampleShutdown_confirmInterrupt`.

Cleanup functions registered with `AddCleanup` run when the context is garbage collected. Unlike handlers they are not triggered by signals — they are a last-resort cleanup path with no ordering guarantees relative to program exit.

```go
ctx.AddCleanup(func() {
    db.Close()
})
```

In `main`, use `defer ctx.Close()` to guarantee cleanup functions run when the function returns. The `ShutdownContext` registers a `runtime.AddCleanup` function that fires when the object is GC'd, but the Go runtime may exit before GC runs at the end of `main`. Calling `Close` explicitly is only necessary there — in all other contexts the runtime cleanup is sufficient.

```go
func main() {
    ctx := context.Shutdown()
    defer ctx.Close()

    ctx.AddCleanup(func() {
        db.Close()
    })
}
```

`Intercept`, `InterceptTimeout`, `AddHandler` and `AddCleanup` are also available as package-level functions that accept any `context.Context`, logging a warning when the context is not a `ShutdownContext`.

Use the package-level `AddCleanup` in constructors that receive a `context.Context` to register cleanup automatically when the caller passes a `ShutdownContext`:

```go
func NewDatabase(ctx context.Context, dsn string) (*Database, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    context.AddCleanup(ctx, func() {
        db.Close()
    })
    return &Database{db: db}, nil
}
```

When `ctx` is a `ShutdownContext` the cleanup is registered; when it is any other context the call is a no-op (with a logged warning), so the constructor works correctly regardless of what the caller provides.


### Full example

```go
package main

import (
	"fmt"
	"time"

	"go.chrisrx.dev/x/context"
)

type DB struct {
	// ...
}

func NewDB(ctx context.Context) *DB {
	context.AddCleanup(ctx, func() {
		fmt.Printf("running database cleanup function\n")
	})
	return &DB{}
}

func main() {
	ctx := context.Shutdown()
	defer ctx.Close()

	ctx.AddHandler(func() {
		fmt.Println("\rCTRL+C pressed, attempting graceful shutdown ...")
		select {} // blocks indefinitely
	})
	ctx.AddHandler(func() {
		fmt.Println("\rCTRL+C pressed again, shutting down immediately ...")
	})
	ctx.AddCleanup(func() {
		fmt.Printf("running cleanup function 1\n")
		time.Sleep(500 * time.Millisecond)
	})
	ctx.AddCleanup(func() {
		fmt.Printf("running cleanup function 2\n")
		time.Sleep(100 * time.Millisecond)
	})

	_ = NewDB(ctx)

	<-ctx.Done()
}
```

<img src="demo.gif" width="600" />
