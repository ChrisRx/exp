package page

import (
	"time"

	"go.chrisrx.dev/x/must"
)

// Offset is used for offset-based pagination.
type Offset struct {
	Token

	// timestamp for initial read
	ReadTimestamp time.Time `$default:"now()"`

	Limit, Offset int
}

func (t Offset) Encode(opts ...Option) string {
	return must.Get0(t.TryEncode(opts...))
}

func (t Offset) TryEncode(opts ...Option) (string, error) {
	return tokenMeta[Offset]{
		Token:   t,
		Version: TokenMetaVersion,
	}.Encode(opts...)
}
