package page

import (
	"fmt"
	"strings"
	"time"

	"go.chrisrx.dev/x/must"
)

// Cursor is used for cursor-based pagination. The keyset is defined in
// [Cursor.After].
type Cursor[T comparable] struct {
	Token

	// timestamp for initial read
	ReadTimestamp time.Time `$default:"now()"`
	// cursor-based values
	After T
	// optional
	Sort []OrderBy
}

func (t Cursor[T]) Encode(opts ...Option) string {
	return must.Get0(t.TryEncode(opts...))
}

func (t Cursor[T]) TryEncode(opts ...Option) (string, error) {
	return tokenMeta[Cursor[T]]{
		Token:   t,
		Version: TokenMetaVersion,
	}.Encode(opts...)
}

func (t Cursor[T]) String() string {
	var sb strings.Builder
	sb.WriteString("Cursor[")
	sb.WriteString(fmt.Sprintf("%+v", t.After))
	if !t.ReadTimestamp.IsZero() {
		sb.WriteString(" read=")
		sb.WriteString(t.ReadTimestamp.Format(time.DateTime))
	}
	sb.WriteString("]")
	return sb.String()
}

type OrderBy struct {
	Column       string
	IsDescending bool
}

func (o OrderBy) String() string {
	if o.IsDescending {
		return o.Column + " DESC"
	}
	return o.Column
}
