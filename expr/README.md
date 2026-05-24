# expr

A Go expression evaluator built on top of the standard library's `go/parser`. Expressions use Go syntax with a handful of ergonomic additions, and can reference a rich set of built-in functions or call methods on values directly.

## Usage

```go
v, err := expr.Eval(`len("hello") > 3`)
// v.Interface() == true

v, err := expr.Eval(`sprintf("%s=%d", "count", 42)`)
// v.Interface() == "count=42"
```

Pass variables into an expression via `Env`:

```go
v, err := expr.Eval(`split_addr(addr).port > 1024`, expr.Env(map[string]reflect.Value{
    "addr": reflect.ValueOf(":8080"),
}))
// v.Interface() == true
```

A special `self` key in the environment is used for implicit single-argument calls:

```go
v, err := expr.Eval(`split_addr().port > 1024`, expr.Env(map[string]reflect.Value{
    "self": reflect.ValueOf(":8080"),
}))
```

## Syntax notes

- Single quotes (`'`) are treated as double quotes, so `'hello'` is a valid string literal.
- Methods on values can be called in snake_case or camelCase — `now().is_zero()` and `now().IsZero()` both work.
- Composite struct literals are supported: `Something{Field: "value"}`.
- Time arithmetic uses `+` / `-` directly: `now() + duration("1h")`.

## Operators

| Category   | Operators                              |
|------------|----------------------------------------|
| Arithmetic | `+` `-` `*` `/` `%`                   |
| Bitwise    | `&` `\|` `^` `<<` `>>`                |
| Comparison | `==` `!=` `<` `>` `<=` `>=`           |
| Logical    | `&&` `\|\|` `!`                        |
| Unary      | `+` `-` `!` `^`                        |

## Built-in functions

### General

| Function | Signature | Description |
|----------|-----------|-------------|
| `len` | `len(v any) int` | Length of a string, slice, array, map, or channel. |
| `some` | `some(v any) bool` | Returns `true` if `v` is non-zero / non-nil. |
| `none` | `none(v any) bool` | Returns `true` if `v` is zero / nil. |

### Type conversion

| Function | Signature | Description |
|----------|-----------|-------------|
| `int` | `int(v any) int` | Converts numeric types or a decimal string to `int`. |
| `float` | `float(v any) float64` | Converts numeric types to `float64`. |
| `string` | `string(v any) string` | Converts any value to its string representation (`fmt.Sprint`). |

### Math

| Function | Signature | Description |
|----------|-----------|-------------|
| `min` | `min(a, b float64) float64` | Smaller of two floats. |
| `max` | `max(a, b float64) float64` | Larger of two floats. |
| `rand` | `rand(args ...any) int` | Random int. No args: unbounded. One arg `n`: `[0, n)`. Two args `min, max`: `[min, max)`. |
| `random` | `random() float64` | Random float64 in `[0.0, 1.0)`. |

### Strings

| Function | Signature | Description |
|----------|-----------|-------------|
| `startswith` | `startswith(s, prefix string) bool` | Reports whether `s` begins with `prefix`. |
| `endswith` | `endswith(s, suffix string) bool` | Reports whether `s` ends with `suffix`. |
| `trim` | `trim(s, cutset string) string` | Removes leading and trailing characters in `cutset`. |
| `upper` | `upper(s string) string` | Converts `s` to upper case. |
| `lower` | `lower(s string) string` | Converts `s` to lower case. |
| `split` | `split(s, sep string) []string` | Splits `s` around each occurrence of `sep`. |
| `atoi` | `atoi(s string) int` | Parses a decimal string as an integer. |
| `itoa` | `itoa(n int) string` | Formats an integer as a decimal string. |
| `quote` | `quote(s string) string` | Returns a Go double-quoted string literal. |
| `unquote` | `unquote(s string) string` | Interprets a Go quoted string literal. |

### Formatting

| Function | Signature | Description |
|----------|-----------|-------------|
| `print` | `print(args ...any)` | `fmt.Print` |
| `printf` | `printf(format string, args ...any)` | `fmt.Printf` |
| `println` | `println(args ...any)` | `fmt.Println` |
| `sprint` | `sprint(args ...any) string` | `fmt.Sprint` |
| `sprintf` | `sprintf(format string, args ...any) string` | `fmt.Sprintf` |
| `sprintln` | `sprintln(args ...any) string` | `fmt.Sprintln` |

### OS / filesystem

| Function | Signature | Description |
|----------|-----------|-------------|
| `getwd` | `getwd() string` | Current working directory. |
| `tempdir` | `tempdir() string` | Default temp directory (`os.TempDir`). |
| `joinpath` | `joinpath(elem ...string) string` | Joins path elements (`filepath.Join`). |
| `getenv` | `getenv(key string) string` | Value of an environment variable. |

### Time

| Function | Signature | Description |
|----------|-----------|-------------|
| `now` | `now() time.Time` | Current local time. |
| `date` | `date(year, month, day[, hour, min, sec, nsec, loc] ...any) time.Time` | Constructs a `time.Time`. Month defaults to `January`, location defaults to `UTC`. |
| `duration` | `duration(s string) time.Duration` | Parses a duration string (e.g. `"1h30m"`). |

### Network

| Function | Signature | Description |
|----------|-----------|-------------|
| `parse_ip` | `parse_ip(s string) net.IP` | Parses an IPv4 or IPv6 address. |
| `parse_mac` | `parse_mac(s string) net.HardwareAddr` | Parses an IEEE 802 MAC address. |
| `split_addr` | `split_addr(s string) {Host string, Port int}` | Splits a `host:port` string into a struct with `Host` and `Port` fields. |

### Hash / crypto

| Function | Signature | Description |
|----------|-----------|-------------|
| `hmac` | `hmac(key, data any) string` | HMAC-SHA256 hex digest. |
| `md5` | `md5(input any) string` | MD5 hex digest. |
| `sha1` | `sha1(input any) string` | SHA-1 hex digest. |
| `sha256` | `sha256(input any) string` | SHA-256 hex digest. |

## Package-qualified functions

In addition to the flat built-ins above, several standard-library packages are available via `<package>.<function>` notation.

### `fmt`

`fmt.Print`, `fmt.Printf`, `fmt.Println`, `fmt.Sprint`, `fmt.Sprintf`

### `math`

`math.Abs`, `math.Acos`, `math.Asin`, `math.Atan`, `math.Ceil`, `math.Cos`, `math.Exp`, `math.Log`, `math.Max`, `math.Min`, `math.Round`, `math.Sin`, `math.Tan`

### `time`

Types / values: `time.Time{}`, `time.Duration(n)`, `time.UTC`, `time.Local`

Functions: `time.Now()`, `time.Date(...)`

Constants: `time.Nanosecond`, `time.Millisecond`, `time.Second`, `time.Minute`, `time.Hour`

```
time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Minute)
```

### `net`

`net.ParseIP`, `net.ParseCIDR`

### `json`

| Function | Description |
|----------|-------------|
| `json.Encode(v any) string` | Marshals `v` to a JSON string. |

### `base64`

| Function | Description |
|----------|-------------|
| `base64.Encode(v any) string` | Standard base64 encoding of a string or `[]byte`. |
| `base64.Decode(s string) string` | Standard base64 decoding. |
