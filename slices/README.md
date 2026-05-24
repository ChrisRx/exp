# slices

Package `slices` extends the standard library [`slices`](https://pkg.go.dev/slices) package with generic functional helpers for transforming, filtering, and working with slices.

**Or** — returns the first non-empty slice:

```go
slices.Or(nil, []string{}, []string{"a", "b"}, []string{"c"})
// ["a", "b"]
```

**Map** — transforms each element using a function:

```go
strs := slices.Map([]int{1, 2, 3}, strconv.Itoa)
// ["1", "2", "3"]
```

**MapErr** — like `Map` but stops at the first error:

```go
vals, err := slices.MapErr([]string{"1", "2", "x"}, strconv.Atoi)
// nil, error (on "x")
```

**MapEntries** — converts a slice to a map:

```go
m := slices.MapEntries([]string{"a", "b", "c"}, func(s string) (string, int) {
    return s, len(s)
})
// map["a":1 "b":1 "c":1]
```

**FlatMap** — maps each element to a slice and concatenates the results:

```go
words := slices.FlatMap([]string{"foo bar", "baz"}, strings.Fields)
// ["foo", "bar", "baz"]
```

**Filter** — keeps elements for which the predicate returns true:

```go
evens := slices.Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
// [2, 4]
```

**FilterMap** — maps and drops zero-value results in one pass:

```go
nonEmpty := slices.FilterMap([]string{"a", "", "b", ""}, func(s string) string { return s })
// ["a", "b"]
```

**FilterMap2** — maps using a `(value, ok)` function and drops elements where `ok` is false:

```go
vals := slices.FilterMap2([]string{"1", "x", "3"}, func(s string) (int, bool) {
    n, err := strconv.Atoi(s)
    return n, err == nil
})
// [1, 3]
```

**Find** — returns the first matching element, or the zero value if none is found:

```go
v := slices.Find([]int{1, 2, 3}, func(n int) bool { return n > 1 })
// 2
```

**Truncate** — returns the slice capped at the given index:

```go
slices.Truncate([]int{1, 2, 3, 4, 5}, 3)
// [1, 2, 3]
```

**N** — returns a slice of integers `[0, n)`:

```go
slices.N(5) // [0, 1, 2, 3, 4]
```

**Partition** — splits a slice into two based on a predicate:

```go
even, odd := slices.Partition([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
// even: [2, 4], odd: [1, 3]
```

**Uniq** — returns only the unique elements, preserving order:

```go
slices.Uniq([]string{"a", "a", "b", "b", "c"})
// ["a", "b", "c"]
```
