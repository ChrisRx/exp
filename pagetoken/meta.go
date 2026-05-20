package pagetoken

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"

	"go.chrisrx.dev/x/convert"
	"go.chrisrx.dev/x/internal/gobx"
	"go.chrisrx.dev/x/options"
)

// TokenMetaVersion versions the data layout for token data.
const TokenMetaVersion = 1

// tokenMeta is a container for a [Token] and any related metadata. It provides
// helper methods for headerless gob encoding of a [Token], and a way to
// sign/verify token data. A version for this meta struct is present in the
// rare case that the layout changes.
type tokenMeta[T Token] struct {
	Token     T
	Signature []byte
	Version   int
}

// TokenMeta wraps a [Token] to allow passing metadata along with a pagination
// token.
func TokenMeta[T Token](token T) *tokenMeta[T] {
	return &tokenMeta[T]{
		Token:   token,
		Version: TokenMetaVersion,
	}
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

// Decode decodes a [Token] from a base64-encoded string.
func (t *tokenMeta[T]) Decode(s string) error {
	t.registerCodec()
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	token, err := convert.Into[tokenMeta[T]](data)
	if err != nil {
		return err
	}
	*t = token
	return nil
}

// Encode encodes a [Token] into a base64-encoded string.
func (t tokenMeta[T]) Encode(opts ...Option) (string, error) {
	t.registerCodec()
	if o := options.New(opts); len(o.key) > 0 {
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

// Sign cryptographically signs and encodes a [Token] into a base64-encoded
// string using HMAC-SHA256.
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

// Verify decodes and cryptographically verifies a [Token] from a
// base64-encoded string using HMAC-SHA256.
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

// Encrypt encrypts and authenticates a [Token] into a base64-encoded string
// using XSalsa20 and Poly1305.
func (t tokenMeta[T]) Encrypt(secret []byte) (string, error) {
	t.registerCodec()
	data, err := convert.Into[[]byte](tokenMeta[T]{
		Token:   t.Token,
		Version: t.Version,
	})
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
func (t *tokenMeta[T]) Decrypt(s string, secret []byte) error {
	t.registerCodec()
	encrypted, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	if len(encrypted) < 24 {
		return fmt.Errorf("insufficient message length")
	}
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	key := sha256.Sum256(secret)
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		return fmt.Errorf("cannot decrypt token")
	}
	token, err := convert.Into[tokenMeta[T]](decrypted)
	if err != nil {
		return err
	}
	*t = token
	return nil
}
