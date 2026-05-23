# pagetoken

Package `pagetoken` provides opaque, base64-encoded pagination tokens for pagination. Tokens are serialized with gob[^1] and optionally encrypted with NaCl secretbox (XSalsa20-Poly1305).

[^1]: The header metadata for gob is excluded to reduce the length of the encoded token. See [internal/gobx](../internal/gobx/codec.go).

## Usage

### Built-in tokens

Three token types are provided out of the box.

**Cursor-based** — encodes a keyset cursor of any comparable type:

```go
type KeySet struct {
    ID        int
    CreatedAt time.Time
}

// Encode
token := pagetoken.Cursor[KeySet]{
    After: KeySet{ID: 123, CreatedAt: time.Now()},
    Sort:  []string{"created_at DESC", "id DESC"},
}
s := token.Encode()

// Parse
parsed, err := pagetoken.ParseCursor[KeySet](s)

// ParseOr
parsed, err := pagetoken.ParseOr(s, pagetoken.Cursor[KeySet]{
    Sort: []string{"id ASC"},
})
```

**Offset-based** — encodes limit/offset values:

```go
token := pagetoken.Offset{Limit: 100, Offset: 200}
s := token.Encode()

// Parse
parsed, err := pagetoken.ParseOffset(s)

// ParseOr
parsed, err := pagetoken.ParseOr(s, pagetoken.Offset{
    Limit: 100,
})
```

**Page-based** — encodes page number and page size:

```go
token := pagetoken.Page{Page: 3, PageSize: 25}
s := token.Encode()

// Parse
parsed, err := pagetoken.ParsePage(s)

// ParseOr
parsed, err := pagetoken.ParseOr(s, pagetoken.Page{
    Page:     1,
    PageSize: 100,
})
```

### ParseOr

`ParseOr` returns a default token when the input string is empty, which is useful for handling the first request in a paginated series:

```go
token, err := pagetoken.ParseOr("", pagetoken.Offset{
    ReadTimestamp: time.Now(),
    Limit:         100,
    Offset:        0,
})
```

### Encrypted tokens

Pass a `Secret` option to encrypt and authenticate the token using XSalsa20-Poly1305. The same secret must be provided to parse.

```go
secret := pagetoken.Secret("my-secret-key")

// Encrypt
token := pagetoken.Cursor[int]{After: 123}
s, err := token.Encrypt(secret)

// Decrypt
parsed, err := pagetoken.ParseCursor[int](s, secret)
```

`Encode`/`Parse`/`ParseOr` also accept the secret as an option:

```go
s := pagetoken.Encode(token, secret)
parsed, err := pagetoken.Parse[pagetoken.Cursor[int]](s, secret)
parsed, err := pagetoken.ParseOr(s, pagetoken.Cusor[int]{},
    After: 123,
}, secret)
```

### Custom tokens

Embed `pagetoken.Token` in any struct to make it usable with `Encode` and `Parse`. The embedded interface acts as a marker — no methods need to be implemented.

```go
type SearchToken struct {
    pagetoken.Token

    Query         string
    Filters       map[string]string
    ReadTimestamp time.Time
    After         int64
}
```

Use the generic `Encode` and `Parse`/`ParseOr` functions directly:

```go
// Encode
token := SearchToken{
    Query:         "go generics",
    Filters:       map[string]string{"lang": "go"},
    ReadTimestamp: time.Now(),
    After:         456,
}
s := pagetoken.Encode(token)

// Parse
parsed, err := pagetoken.Parse[SearchToken](s)

// ParseOr
parsed, err := pagetoken.ParseOr[SearchToken](s, SearchToken{
    ReadTimestamp: time.Now(),
})
```

Encryption works the same way for custom tokens:

```go
secret := pagetoken.Secret("my-secret-key")

s, err := pagetoken.TryEncode(token, secret)

parsed, err := pagetoken.Parse[SearchToken](s, secret)
```
