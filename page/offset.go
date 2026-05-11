package page

import (
	"strconv"
	"strings"
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

func (t Offset) String() string {
	var sb strings.Builder
	sb.WriteString("Offset[")
	sb.WriteString("limit=")
	sb.WriteString(strconv.Itoa(t.Limit))
	sb.WriteString(" offset=")
	sb.WriteString(strconv.Itoa(t.Offset))
	if !t.ReadTimestamp.IsZero() {
		sb.WriteString(" read=")
		sb.WriteString(t.ReadTimestamp.Format(time.DateTime))
	}
	sb.WriteString("]")
	return sb.String()
}
