package pagetoken

import (
	"go.chrisrx.dev/x/options"
)

type TokenOptions struct {
	key    []byte
	secret []byte
}

type Option = options.Option[*TokenOptions]

// Signed is an option that specifies a key to be used for signing tokens.
type Signed []byte

func (s Signed) Apply(o *TokenOptions) {
	o.key = s
}

// Secret is an option that specifies a secret to be used for
// encryption/decryption.
type Secret []byte

func (s Secret) Apply(o *TokenOptions) {
	o.secret = s
}

// Parse parses a [Token] from a base64-encoded string, using any provided
// options.
func Parse[T Token](s string, opts ...Option) (T, error) {
	if s == "" {
		return *new(T), nil
	}
	o := options.New(opts)
	var token tokenMeta[T]
	if len(o.secret) > 0 {
		if err := token.Decrypt(s, o.secret); err != nil {
			return *new(T), err
		}
		return token.Token, nil
	}
	if err := token.Decode(s); err != nil {
		return *new(T), err
	}
	if len(o.key) > 0 && len(token.Signature) > 0 {
		if err := token.Verify(o.key); err != nil {
			return *new(T), err
		}
	}
	return token.Token, nil
}

// ParseOr parses a [Token] from a base64-encoded string. When the string is
// empty, the provided default token is returned.
func ParseOr[T Token](s string, defaultPageToken T, opts ...Option) (T, error) {
	if s != "" {
		return Parse[T](s, opts...)
	}
	return defaultPageToken, nil
}
