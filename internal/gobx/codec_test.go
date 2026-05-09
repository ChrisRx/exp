package gobx_test

import (
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/internal/gobx"
	"go.chrisrx.dev/x/must"
)

func TestCodec(t *testing.T) {
	type Token struct {
		Limit, Offset int64
	}
	var enc gobx.Codec[Token]
	expected := Token{Limit: 100, Offset: 100}
	data, err := enc.Encode(expected)
	assert.NoError(t, err)
	token, err := enc.Decode(data)
	assert.NoError(t, err)
	assert.Equal(t, expected, token)

	assert.Error(t,
		"must provide non-pointer type",
		must.Get1(new(gobx.Codec[*Token]).Encode(&token)),
	)
}
