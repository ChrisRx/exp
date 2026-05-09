package page

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"go.chrisrx.dev/x/convert"
	"go.chrisrx.dev/x/internal/gobx"
	"go.chrisrx.dev/x/structs"
)

type options struct {
	key []byte
}

type Option func(*options)

func Signed(key []byte) Option {
	return func(o *options) {
		o.key = key
	}
}

type Token interface {
	isToken()
}

// NewToken constructs a new token with default values.
func NewToken[T Token]() T {
	return structs.DefaultsFor[T]()
}

// ParseToken parses an encoded token from a string.
func ParseToken[T Token](s string, opts ...Option) (T, error) {
	if s == "" {
		return NewToken[T](), nil
	}
	var token tokenMeta[T]
	if err := token.Decode(s, opts...); err != nil {
		return *new(T), err
	}
	return token.Token, nil
}

// TokenMetaVersion versions the data layout for token data.
const TokenMetaVersion = 1

type tokenMeta[T any] struct {
	Token     T
	Signature []byte
	Version   int
}

func (t tokenMeta[T]) registerCodec() {
	if _, ok := convert.LookupFor[tokenMeta[T], []byte](); !ok {
		var codec gobx.Codec[tokenMeta[T]]
		convert.Register(func(data []byte, opts ...convert.Option) (tokenMeta[T], error) {
			return codec.Decode(data)
		})
		convert.Register(func(v tokenMeta[T], opts ...convert.Option) ([]byte, error) {
			return codec.Encode(v)
		})
	}
}

func (t *tokenMeta[T]) Decode(s string, opts ...Option) error {
	t.registerCodec()
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	token, err := convert.Into[tokenMeta[T]](data)
	if err != nil {
		return err
	}
	if len(o.key) > 0 && len(token.Signature) > 0 {
		if err := token.Verify(o.key); err != nil {
			return err
		}
	}
	*t = token
	return nil
}

func (t tokenMeta[T]) Encode(opts ...Option) (string, error) {
	t.registerCodec()
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if len(o.key) > 0 {
		if err := t.Sign(o.key); err != nil {
			return "", err
		}
	}
	data, err := convert.Into[[]byte](t)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (t *tokenMeta[T]) Sign(key []byte) error {
	t.registerCodec()
	data, err := convert.Into[[]byte](tokenMeta[T]{
		Token:   t.Token,
		Version: t.Version,
	})
	if err != nil {
		return err
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	t.Signature = h.Sum(nil)
	return nil
}

func (t tokenMeta[T]) Verify(key []byte) error {
	t.registerCodec()
	if len(t.Signature) == 0 {
		return fmt.Errorf("signature not found")
	}
	if len(t.Signature) < sha256.Size {
		return fmt.Errorf("insufficient message length")
	}
	data, err := convert.Into[[]byte](tokenMeta[T]{
		Token:   t.Token,
		Version: t.Version,
	})
	if err != nil {
		return err
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	if !hmac.Equal(t.Signature, h.Sum(nil)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
