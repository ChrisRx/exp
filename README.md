# exp

[![Go Reference](https://pkg.go.dev/badge/go.chrisrx.dev/x.svg)](https://pkg.go.dev/go.chrisrx.dev/x)
[![Build Status](https://github.com/ChrisRx/exp/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ChrisRx/exp/actions)

Experimental Go packages being evaluated for reusability and stability. All packages share a single Go module and are imported as `go.chrisrx.dev/x/<subpackage>`.

When a package's API stabilizes and proves useful, it is promoted to its own module:

```go
// before promotion
import "go.chrisrx.dev/x/env"

// after promotion
import "go.chrisrx.dev/env"
```

## Extended standard library packages

Several packages in this module are designed as **drop-in replacements** for their standard library counterparts. Swapping the import path gives you all the original stdlib functions plus additional utilities — no other code changes required.

| This module | Extends |
|-------------|---------|
| `go.chrisrx.dev/x/context` | `context` |
| `go.chrisrx.dev/x/slices` | `slices` |
| `go.chrisrx.dev/x/strings` | `strings` |
| `go.chrisrx.dev/x/errors` | `errors` |
| `go.chrisrx.dev/x/sync` | `sync` |

This is implemented with [`aliaspkg`](https://pkg.go.dev/go.chrisrx.dev/tools#readme-aliaspkg), a code-generation tool that re-exports every symbol from the target stdlib package under the replacement package name. The generated file is committed as `alias.go` and is never edited by hand. Each re-exported function includes a doc comment linking back to the original stdlib entry so godoc references remain intact.

```go
// Before: only stdlib slices functions available.
import "slices"

// After: all stdlib functions plus Filter, Map, FlatMap, Find, Partition, Uniq, etc.
import "go.chrisrx.dev/x/slices"
```

To regenerate the alias files for packages after a Go version upgrade, run:

```
task generate
```

## Highlights

### slices

A superset of the standard library [`slices`](https://pkg.go.dev/slices) package. It can be used as a drop-in replacement — see [Extended standard library packages](#extended-standard-library-packages) — while adding generic higher-order functions missing from the stdlib:

```go
import "go.chrisrx.dev/x/slices"

names := []string{"alice", "bob", "chris"}

// Filter keeps elements matching a predicate.
long := slices.Filter(names, func(s string) bool { return len(s) > 3 })

// Map transforms elements.
upper := slices.Map(names, strings.ToUpper)

// Partition splits into two slices based on a predicate.
short, long := slices.Partition(names, func(s string) bool { return len(s) <= 3 })

// Uniq removes duplicates (works on any type, including uncomparable ones).
unique := slices.Uniq([]string{"a", "b", "a", "c"})

// FlatMap maps then flattens.
words := slices.FlatMap(sentences, strings.Fields)
```

See [slices/README.md](slices/README.md).

### set

A generic set that works with **any** type, including uncomparable types like `[]byte`, using [`maphash`](https://pkg.go.dev/hash/maphash) internally. The zero value is ready to use.

```go
import "go.chrisrx.dev/x/set"

// Works with comparable types.
s := set.New(1, 2, 2, 3, 3, 3)
fmt.Println(s.List()) // [1 2 3]

// Also works with uncomparable types.
var bs set.Set[[]byte]
bs.Add([]byte("foo"), []byte("foo"), []byte("bar"))
fmt.Println(bs.Len()) // 2

// Set operations.
a := set.New("x", "y")
b := set.New("y", "z")
fmt.Println(a.Union(b).List())        // [x y z]
fmt.Println(a.Intersection(b).List()) // [y]
fmt.Println(a.Difference(b).List())   // [x]
```

### env

Populates a struct from environment variables using struct tags. Supports default values (including Go expression defaults via `$default`), expression-based validation, automatic prefix derivation from struct nesting, and custom type parsers.

```go
import "go.chrisrx.dev/x/env"

cfg := env.MustParseFor[struct {
    Addr           string        `env:"ADDR"           default:":8080"  validate:"split_addr(self).port > 1024"`
    ReadTimeout    time.Duration `env:"TIMEOUT"        default:"2m"`
    MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" $default:"1 << 20"`
}](env.RootPrefix("MYSERVICE"))
// Reads MYSERVICE_ADDR, MYSERVICE_TIMEOUT, MYSERVICE_MAX_HEADER_BYTES.
```

Nested structs automatically extend the prefix chain, so each level of nesting maps cleanly to a `PARENT_CHILD_FIELD` naming scheme without any manual configuration.

### group

Manages a pool of goroutines with optional bounded concurrency, context propagation, and first-error cancellation. `ResultGroup` extends this to collect typed return values as an `iter.Seq2` iterator, making fan-out patterns concise.

```go
import "go.chrisrx.dev/x/group"

// Basic fan-out — cancel all goroutines on first error.
g := group.New(ctx, group.WithLimit(8))
for _, item := range items {
    g.Go(func(ctx context.Context) error {
        return process(ctx, item)
    })
}
if err := g.Wait(); err != nil {
    log.Fatal(err)
}

// Collect typed results as an iterator.
rg := group.NewResultGroup[string](ctx, group.WithLimit(4))
for _, url := range urls {
    rg.Go(func(ctx context.Context) (string, error) {
        return fetch(ctx, url)
    })
}
for result, err := range rg.Get() {
    if err != nil { ... }
    fmt.Println(result)
}
```

### run

Retry and polling primitives built on exponential backoff. `Every` runs a function on a fixed interval until the context is cancelled. `Until` and `Unless` retry until a condition is met or fails. All three accept either a `time.Duration` or a full `Options` value for fine-grained backoff control.

```go
import "go.chrisrx.dev/x/run"

// Call a function every 500ms until the context is done.
run.Every(ctx, func() {
    metrics.Collect()
}, 500*time.Millisecond)

// Retry until the function returns nil (no error).
err := run.Until(ctx, func() error {
    return db.Ping()
}, run.Options{
    InitialInterval: 100 * time.Millisecond,
    MaxInterval:     5 * time.Second,
    MaxAttempts:     10,
})

// Retry until the function returns false.
err = run.Unless(ctx, func() bool {
    return queue.Full()
}, time.Second)
```

### must

Unwraps `(T, error)` return values and panic-recovery helpers. `must.Ok` panics on a non-nil error, eliminating boilerplate in initialization code. `must.Catch` converts a panic back into an error for deferred recovery. `must.Recover` silently absorbs panics (optionally only specific ones) and logs them.

```go
import "go.chrisrx.dev/x/must"

// Unwrap (value, error) — panics if err != nil.
f := must.Ok(os.Open("config.yaml"))

// Catch a panic and surface it as an error return.
func parse(data []byte) (result Result, err error) {
    defer must.Catch(&err)
    result = mustParse(data) // panics on bad input
    return
}

// Recover all panics, logging them via slog.
defer must.Recover()

// Recover only a specific error.
defer must.Recover(io.ErrUnexpectedEOF)
```

### safe

Converts panics to errors and provides struct-level compile-time guards. `safe.Do` wraps any `func()` and returns a panic as an error. The marker types `NoCompare`, `NoCopy`, and `NoTypeConversion` are zero-size fields that enforce invariants at compile time or via `go vet`.

```go
import "go.chrisrx.dev/x/safe"

// Run untrusted code without crashing the process.
err := safe.Do(func() {
    riskyOperation()
})

// Prevent a struct from being copied after first use (triggers go vet).
type Buffer struct {
    _ safe.NoCopy
    // ...
}

// Prevent unsafe type conversion between two struct types.
type MyID struct {
    _ safe.NoTypeConversion[MyID]
    v uint64
}
```

### log

Wraps `log/slog` with environment-driven configuration. Call `log.SetDefault()` once at startup to read `LOG_LEVEL`, `LOG_FORMAT` (`text`/`json`), `LOG_ADD_SOURCE`, and `LOG_REMOVE_ATTRS` from the environment and install a configured default logger. `log.New` builds a standalone `*slog.Logger` from the same option set.

```go
import "go.chrisrx.dev/x/log"

func main() {
    // Configure the default slog.Logger from environment variables once.
    log.SetDefault()

    // Build a logger with explicit options (overrides env defaults).
    logger := log.New(
        log.WithFormat(log.JSONFormat),
        log.WithLevel(slog.LevelDebug),
    )

    // Fatal variants log via slog then call os.Exit(1).
    // Set GO_BACKTRACE=1 to print the full stack trace to stderr.
    log.Fatalf("startup failed: %v", err)
}
```

### context

A drop-in replacement for the standard library [`context`](https://pkg.go.dev/context) package, adding type-safe generic keys and signal-based graceful shutdown.

```go
import "go.chrisrx.dev/x/context"

// Type-safe key — no unexported struct types or manual type assertions needed.
var AuthKey = context.Key[*Auth]()

ctx = AuthKey.WithValue(ctx, &Auth{Token: "Bearer ..."})
auth := AuthKey.Value(ctx)   // zero value if not set
AuthKey.Has(ctx)             // true/false
AuthKey.ValueFunc(ctx, func(a *Auth) { ... }) // called only when present

// Signal-based shutdown — handlers run one per signal, in registration order.
ctx := context.Shutdown()

ctx.AddHandler(func() {
    fmt.Println("draining...")
    server.Shutdown(context.Background())
})
ctx.AddHandler(func() {
    os.Exit(1) // second signal: force exit
})

ctx.Wait() // blocks until all handlers have run
```

## Packages

| Package | Description |
|---------|-------------|
| [assert](assert) | Testing assertions for values, errors, and eventual conditions with diff output |
| [backoff](backoff) | Exponential backoff with configurable intervals, multipliers, and jitter |
| [chans](chans) | Channel utilities: collect, drain, and limit operations |
| [constraints](constraints) | Generic numeric constraint types (`Signed`, `Unsigned`, `Integer`, `Float`, `Complex`) |
| [context](context/README.md) | Drop-in `context` replacement with type-safe generic keys and signal-based graceful shutdown |
| [convert](convert) | Type-safe conversions via a registry of conversion functions |
| [env](env/README.md) | Parse environment variables into struct fields using tags |
| [errors](errors) | Drop-in `errors` replacement with generic `As`, `Wrap`, `Ignore`, and `Stack` |
| [expr](expr/README.md) | Parse and evaluate Go expressions at runtime using reflection |
| [future](future) | Lazy-evaluated values with channel-based synchronization |
| [group](group) | Goroutine pool with concurrency limits, context cancellation, and error handling |
| [log](log) | `slog.Logger` construction with environment-based configuration |
| [maps](maps) | Map utilities: filter, transform, and convert to slices |
| [math](math) | Generic math functions (e.g. `Sum`) for numeric types |
| [must](must) | Unwrap error tuples, panicking on non-nil errors |
| [options](options) | Generic option pattern with composable `Option[T]` interface |
| [pagetoken](pagetoken/README.md) | Encrypted, base64-serialized pagination tokens |
| [ptr](ptr) | Pointer helpers: create, dereference, nullable conversion, zero-value check |
| [random](random) | Cryptographically secure random string generation |
| [result](result) | `Of[T]` result type with `iter.Seq` support |
| [run](run) | Retry with configurable exponential backoff, max attempts, and time limits |
| [safe](safe) | Execute functions safely, converting panics to errors |
| [set](set) | Generic set supporting uncomparable values via `maphash` |
| [slices](slices/README.md) | Slice utilities: filter, map, find, truncate |
| [sort](sort) | Sorted map view with pagination and iteration |
| [stack](stack) | Caller info from the runtime, filtering internal and stdlib frames |
| [strings](strings) | String utilities: dedent, title-case, case conversion, whitespace handling |
| [structs](structs) | Struct field parsing and default-value application via tags |
| [sync](sync) | Synchronization primitives: `Chan`, `Semaphore`, `WaitGroup`, `Once` |
| [uuid](uuid) | Deterministic RFC 4122 v8 UUIDs from strings using FNV-1a |
