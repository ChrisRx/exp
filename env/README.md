# env

[![Go Reference](https://pkg.go.dev/badge/go.chrisrx.dev/x.svg)](https://pkg.go.dev/go.chrisrx.dev/x/env)
[![Build Status](https://github.com/ChrisRx/exp/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ChrisRx/exp/actions)

A library for declaring composable, reusable Go structs that loads values parsed from environment variables.

## 🚀 Features

* Supports many [commonly used types](#supported-types)
* Register [user-defined types](#registering-custom-parsers)
* Nested structs auto-prefix environment variable keys
* Use Go-like expressions to generate default values

> [!IMPORTANT]
> Features like auto-prefix are important to making structs composable, which is why it is on by default.

## 📋 Usage

```go
var opts = env.MustParseAs[struct {
    Timeout time.Duration `env:"TIMEOUT" default:"10m"`
    Start   time.Time     `env:"START" $default:"now()"`
}]()


func main() {
    ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
    defer cancel()

    fmt.Printf("started at %v\n", opts.Start)
}
```

See [testdata/pg](testdata/pg/config.go) for a more complete example of a reusable configuration struct.

### Supported types

> [!NOTE]
> Pointers to any types listed will also work.

Basic types:
* `string`, `[]byte`
* `int`, `int8`, `int16`, `int32`, `int64`
* `uint`, `uint8`, `uint16`, `uint32`, `uint64`
* `float32`, `float64`
* `bool`

It also supports maps, slices, and nested structs.

> [!CAUTION]
> Slices or maps do not support nesting other slices/maps. This limitation is mostly due to needing a way to split string input into elements becomes difficult to express in struct tags beyond one level. This could change if an ergonomic solution is found that keeps configuration still relatively simple.

Built-in customer parsers:
* `time.Time`, `time.Duration`
* `url.URL`
* `rsa.PublicKey`
* `x509.Certificate`
* `net.IP`

Any existing type that implements `encoding.TextUnmarshaler` will also work.

### Registering custom parsers

The [Register](https://pkg.go.dev/go.chrisrx.dev/x/env#Register) function can be used to define custom type parsers. It takes a non-pointer type parameter for the custom type and the parser function as the argument:

```go
env.Register[net.IP](func(field Field, s string) (any, error) {
    return net.ParseIP(s), nil
})
```

The type parameter must be a non-pointer, but registering a type will always work with both the provided type and the pointer version without needing to register them both.
