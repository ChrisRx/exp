package pagetoken

import (
	"time"
)

// Cursor is used for cursor-based pagination. The keyset is defined in
// [Cursor.After].
type Cursor[T comparable] struct {
	Token

	// timestamp for initial read
	ReadTimestamp time.Time
	// cursor-based values
	After T
	// order by
	Sort []string

	// last encoding error
	err error
}

// ParseCursor parses a [Cursor] token from a base64-encoded string.
func ParseCursor[T comparable](s string, opts ...Option) (Cursor[T], error) {
	return Parse[Cursor[T]](s, opts...)
}

// Encode encodes the current token values. If it encounters an error, it
// returns an empty string and provides the error via [Cursor.Err].
//
// There is effectively no need to check the error after successful encoding,
// so it can be skipped in favor of better ergonomics.
func (t Cursor[T]) Encode() string {
	s, err := TryEncode(t)
	if err != nil {
		t.err = err
	}
	return s
}

// Encrypt encodes and encrypts the current token values.
func (t Cursor[T]) Encrypt(secret []byte) (string, error) {
	return Encrypt(t, secret)
}

// Err returns the last error that was encountered while encoding.
func (t Cursor[T]) Err() error {
	return t.err
}

// Offset is used for offset-based pagination.
type Offset struct {
	Token

	// timestamp for initial read
	ReadTimestamp time.Time
	// limit/offset
	Limit, Offset int
	// order by
	Sort []string

	err error
}

// ParseOffset parses an [Offset] token from a base64-encoded string.
func ParseOffset(s string, opts ...Option) (Offset, error) {
	return Parse[Offset](s, opts...)
}

// Encode encodes the current token values. If it encounters an error, it
// returns an empty string and provides the error via [Offset.Err].
//
// There is effectively no need to check the error after successful encoding,
// so it can be skipped in favor of better ergonomics.
func (t Offset) Encode() string {
	s, err := TryEncode(t)
	if err != nil {
		t.err = err
	}
	return s
}

// Encrypt encodes and encrypts the current token values.
func (t Offset) Encrypt(secret []byte) (string, error) {
	return Encrypt(t, secret)
}

// Err returns the last error that was encountered while encoding.
func (t Offset) Err() error {
	return t.err
}

// Page is used for page-based pagination.
type Page struct {
	Token

	// timestamp for initial read
	ReadTimestamp time.Time
	// page number and size
	Page, PageSize int
	// order by
	Sort []string

	err error
}

// ParsePage parses a [Page] token from a base64-encoded string.
func ParsePage(s string, opts ...Option) (Page, error) {
	return Parse[Page](s, opts...)
}

// Encode encodes the current token values. If it encounters an error, it
// returns an empty string and provides the error via [Page.Err].
//
// There is effectively no need to check the error after successful encoding,
// so it can be skipped in favor of better ergonomics.
func (t Page) Encode() string {
	s, err := TryEncode(t)
	if err != nil {
		t.err = err
	}
	return s
}

// Encrypt encodes and encrypts the current token values.
func (t Page) Encrypt(secret []byte) (string, error) {
	return Encrypt(t, secret)
}

// Err returns the last error that was encountered while encoding.
func (t Page) Err() error {
	return t.err
}
