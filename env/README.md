# env

[![Go Reference](https://pkg.go.dev/badge/go.chrisrx.dev/x.svg)](https://pkg.go.dev/go.chrisrx.dev/x/env)
[![Build Status](https://github.com/ChrisRx/exp/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/ChrisRx/exp/actions)

A library for declaring composable, reusable Go structs for loading values parsed from environment variables.

## 🚀 Features

* Supports many [commonly used types](#supported-types)
* Register [user-defined types](#registering-custom-parsers)
* Nested structs [auto-prefix](#auto-prefix) environment variable keys
* Use [Go-like expressions](../expr/README.md) to [generate default values](#default-expressions) or [validate values](#field-validation)

> [!IMPORTANT]
> Features like auto-prefix are important to making structs composable, which is why it is on by default.

## 📋 Usage

```go
var opts = env.MustParseFor[struct {
	Addr           string        `env:"ADDR" default:":8080" validate:"int(split(Addr, ':')[1]) > 1024"`
	Dir            http.Dir      `env:"DIR" $default:"tempdir()"`
	ReadTimeout    time.Duration `env:"TIMEOUT" default:"2m"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" $default:"1 << 20"`
}](env.Namespace("FILESERVER"))

func main() {
	s := &http.Server{
		Addr:           opts.Addr,
		Handler:        http.FileServer(opts.Dir),
		ReadTimeout:    opts.ReadTimeout,
		WriteTimeout:   opts.WriteTimeout,
		MaxHeaderBytes: opts.MaxHeaderBytes,
	}
	log.Printf("serving %s at %s ...\n", opts.Dir, opts.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

See [testdata/pg](testdata/pg/config.go) for a more complete example of a reusable configuration struct.

### Supported types

Basic types:
* `string`, `[]byte`
* `int`, `int8`, `int16`, `int32`, `int64`
* `uint`, `uint8`, `uint16`, `uint32`, `uint64`
* `float32`, `float64`
* `bool`

Go slices, maps and structs are also supported.

> [!WARNING]
> Slices and maps cannot nest slices, maps or structs.

Built-in custom parsers:
* `time.Time`, `time.Duration`
* `url.URL`
* `rsa.PublicKey`
* `x509.Certificate`
* `net.IP`

Any existing type that implements `encoding.TextUnmarshaler` will also work.

> [!NOTE]
> Pointers to any types listed will also work.

### Tags

The following struct tags are used to define how env reads from environment variables into values:

| Name        | Description |
| ----------- | ----------- |
| `env`       | The name of the environment variable to load values from. If not present, the field is ignored. |
| `default`   | Specifies a default value for a field. This is used when the environment variable is not set. |
| `$default`  | Use an expression to set a default field value. This is used when the environment variable is not set. |
| `validate`  | Use a boolean expression to validate the field value. |
| `required`  | Set the field as required.  |
| `sep`       | Separator used when parsing array/slice values. Defaults to `,`. |
| `layout`    | Layout used to format/parse `time.Time` fields. Defaults to [time.RFC3339Nano](https://pkg.go.dev/time#RFC3339Nano). |

### Auto-prefix

Nested structs will automatically prepend the field name to the environment variable name for nested fields:

```go
type Config struct {
    DB struct {
        Host string `env:"HOST"`
        Port int    `env:"PORT"`
    }
}
```

The above struct has a nested anonymous struct with the field name `DB`, which results in the nested fields values being loaded from `DB_HOST` and `DB_PORT`.

The prefix used for a struct can be set explicitly using the `env` tag on the struct itself:

```go
type Config struct {
    DB struct {
        Host string `env:"HOST"`
        Port int    `env:"PORT"`
    } `env:"USERS_DB"`
}
```

Setting `env:"USERS_DB"` here means that the environment variables are now loaded from `USERS_DB_HOST`/`USERS_DB_PORT`.

> [!TIP]
> You can prevent auto-prefix on a nested struct by declaring it as an anonymous field, aka embedding.

### Registering custom parsers

The [Register](https://pkg.go.dev/go.chrisrx.dev/x/env#Register) function can be used to define custom type parsers. It takes a non-pointer type parameter for the custom type and the parser function as the argument:

```go
env.Register[net.IP](func(field Field, s string) (any, error) {
    return net.ParseIP(s), nil
})
```

The type parameter must be a non-pointer, but registering a type will always work with both the provided type and the pointer version without needing to register them both.

### Default expressions

Default values can be generated using a Go-like expression language:

```go
type Config struct {
    Start time.Time `env:"START" $default:"now()"`
}
```

Unlike `default`, the `$default` tag will always evaluate the tag value as an expression. There are quite a few [builtins available](../expr/README.md#builtins) that can be used and custom ones can be added. Functions like `now()` return a [time.Time](https://pkg.go.dev/time#Time) so that method chaining can be used to construct more complex expressions:

```go
type Config struct {
    End time.Time `env:"END" $default:"now().add(duration('1h'))"`
}
```

A couple interesting things are happening here. For one, the above is syntactic sugar for `time.Now().Add(-1 * time.Hour)`. It works by transforming the method lookups for Go types from snakecase to the expected Go method text case. For example, `now().is_zero()` will call `time.Now().IsZero()`, under-the-hood.

The other interesting thing happening here is that strings are specified using single quotes. This was a workaround to deal with the strict requirements for parsing [struct tags](https://pkg.go.dev/reflect#StructTag) which requires using double quotes to enclose tag values.

### Field validation

The `validate` tag can be used to specify a boolean expression that checks the value of a field once parsing is finished. This can be used to verify things like minimum string length:

```go
type Config struct {
    Name string `env:"NAME" validate:"len(Name) > 3"`
}
```

The field value is injected into the expression scope allowing for it to be referenced by the field name.
