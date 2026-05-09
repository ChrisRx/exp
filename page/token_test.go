package page_test

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/page"
)

func TestToken(t *testing.T) {
	t.Run("cursor", func(t *testing.T) {
		now := time.Now()

		token := page.NewToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]]()
		token.After.ID = 123
		token.After.CreatedAt = now

		s, err := token.TryEncode()
		assert.NoError(t, err)

		parsed, err := page.ParseToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]](s)
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 123, token.After.ID)
		assert.Equal(t, now, token.After.CreatedAt)

		token.After.ID = 124

		s, err = token.TryEncode()
		assert.NoError(t, err)
		parsed, err = page.ParseToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]](s)
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 124, token.After.ID)
		assert.Equal(t, now, token.After.CreatedAt)
	})

	t.Run("offset", func(t *testing.T) {
		token := page.NewToken[page.Offset]()
		token.Limit = 100
		token.Offset = 200

		s, err := token.TryEncode()
		assert.NoError(t, err)

		parsed, err := page.ParseToken[page.Offset](s)
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 100, token.Limit)
		assert.Equal(t, 200, token.Offset)
	})

	t.Run("signed", func(t *testing.T) {
		key := []byte("secret")

		token := page.NewToken[page.Cursor[int]]()
		token.After = 123
		s := token.Encode(page.Signed(key))
		parsed, err := page.ParseToken[page.Cursor[int]](s, page.Signed(key))
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		parsed, err = page.ParseToken[page.Cursor[int]](s, page.Signed([]byte("wrong key")))
		assert.Error(t, "signature mismatch", err)
	})
}

func BenchmarkToken(b *testing.B) {
	b.Run("encode", func(b *testing.B) {
		token := page.NewToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]]()
		token.After.ID = 123
		b.ResetTimer()
		for b.Loop() {
			token.Encode()
		}
	})

	b.Run("sign", func(b *testing.B) {
		token := page.NewToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]]()
		b.ResetTimer()
		for b.Loop() {
			token.Encode(page.Signed([]byte("secret")))
		}
	})

	b.Run("parse", func(b *testing.B) {
		token := page.NewToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]]()
		s := token.Encode()
		b.ResetTimer()
		for b.Loop() {
			_, _ = page.ParseToken[page.Cursor[struct {
				ID        int
				CreatedAt time.Time
			}]](s)
		}
	})

	b.Run("parse signed", func(b *testing.B) {
		token := page.NewToken[page.Cursor[struct {
			ID        int
			CreatedAt time.Time
		}]]()
		key := page.Signed([]byte("secret"))
		s := token.Encode(key)
		b.ResetTimer()
		for b.Loop() {
			_, _ = page.ParseToken[page.Cursor[struct {
				ID        int
				CreatedAt time.Time
			}]](s, key)
		}
	})
}
