package pagetoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"

	"go.chrisrx.dev/x/convert"
	"go.chrisrx.dev/x/internal/gobx"
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/options"
)

type tokenOptions struct {
	secret []byte
}

type Option = options.Option[*tokenOptions]

// Secret is an option that specifies a secret to be used for
// encryption/decryption.
type Secret []byte

func (s Secret) Apply(o *tokenOptions) {
	o.secret = s
}

// The Token interface is embedded in structs intended to be used as pagination
// tokens with this package. This helps ensure that custom tokens are only
// structs.
type Token interface {
	isToken()
}

// Encode encodes a [Token] into a base64-encoded string. It panics if
// encountering an error during encoding.
//
// This panics instead of returning an error to improve ergonomics. After a
// type is successfully encoded once, subsequent calls will only produce an
// error in the most extreme situations (e.g. https://go.dev/issue/66821). If
// there are any concerns with this behavior [TryEncode] can be used instead.
func Encode[T Token](token T, opts ...Option) string {
	return must.Ok(TryEncode(token, opts...))
}

// TryEncode encodes a [Token] into a base64-encoded string.
func TryEncode[T Token](token T, opts ...Option) (string, error) {
	registerCodec[T]()
	switch o := options.New(opts); {
	case len(o.secret) > 0:
		return Encrypt(token, o.secret)
	}
	data, err := convert.Into[[]byte](token)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func registerCodec[T any]() {
	if _, ok := convert.LookupFor[T, []byte](); !ok {
		var codec gobx.Codec[T]
		convert.Register(func(data []byte, opts ...convert.Option) (T, error) {
			return codec.Decode(data)
		})
		convert.Register(func(v T, opts ...convert.Option) ([]byte, error) {
			return codec.Encode(v)
		})
	}
}

// Parse parses a [Token] from a base64-encoded string, using any provided
// options.
func Parse[T Token](s string, opts ...Option) (T, error) {
	if s == "" {
		return *new(T), nil
	}
	switch o := options.New(opts); {
	case len(o.secret) > 0:
		return Decrypt[T](s, o.secret)
	default:
		registerCodec[T]()
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return *new(T), err
		}
		token, err := convert.Into[T](data)
		if err != nil {
			return *new(T), err
		}
		return token, nil
	}
}

// ParseOr parses a [Token] from a base64-encoded string. When the string is
// empty, the provided default token is returned.
func ParseOr[T Token](s string, defaultPageToken T, opts ...Option) (T, error) {
	if s != "" {
		return Parse[T](s, opts...)
	}
	return defaultPageToken, nil
}

// Encrypt encrypts and authenticates a [Token] into a base64-encoded string
// using XSalsa20 and Poly1305.
func Encrypt[T Token](token T, secret []byte) (string, error) {
	registerCodec[T]()
	data, err := convert.Into[[]byte](token)
	if err != nil {
		return "", err
	}
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}
	key := sha256.Sum256(secret)
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt authenticates and decrypts a [Token] from a base64-encoded string
// using XSalsa20 and Poly1305.
func Decrypt[T Token](s string, secret []byte) (T, error) {
	registerCodec[T]()
	encrypted, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return *new(T), err
	}
	if len(encrypted) < 24 {
		return *new(T), fmt.Errorf("insufficient message length")
	}
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	key := sha256.Sum256(secret)
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		return *new(T), fmt.Errorf("cannot decrypt token")
	}
	return convert.Into[T](decrypted)
}
