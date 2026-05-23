package pagetoken_test

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/pagetoken"
)

type KeySet struct {
	ID        int
	CreatedAt time.Time
}

type Token = pagetoken.Cursor[KeySet]

func TestToken(t *testing.T) {
	t.Run("cursor", func(t *testing.T) {
		now := time.Now()

		token := Token{
			After: KeySet{
				ID:        123,
				CreatedAt: now,
			},
		}

		parsed, err := pagetoken.ParseCursor[KeySet](token.Encode())
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 123, token.After.ID)
		assert.Equal(t, now, token.After.CreatedAt)

		token.After.ID = 124

		parsed, err = pagetoken.ParseCursor[KeySet](token.Encode())
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 124, token.After.ID)
		assert.Equal(t, now, token.After.CreatedAt)
	})

	t.Run("offset", func(t *testing.T) {
		token := pagetoken.Offset{
			Limit:  100,
			Offset: 200,
		}
		parsed, err := pagetoken.ParseOffset(token.Encode())
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		assert.Equal(t, 100, token.Limit)
		assert.Equal(t, 200, token.Offset)
	})

	t.Run("encrypted", func(t *testing.T) {
		secret := pagetoken.Secret("secret")

		token := pagetoken.Cursor[int]{
			After: 123,
		}
		encrypted, err := token.Encrypt(secret)
		assert.NoError(t, err)
		parsed, err := pagetoken.ParseCursor[int](encrypted, secret)
		assert.NoError(t, err)
		parsed, err = pagetoken.Decrypt[pagetoken.Cursor[int]](encrypted, secret)
		assert.NoError(t, err)
		assert.Equal(t, token, parsed)
		parsed, err = pagetoken.ParseCursor[int](encrypted, pagetoken.Secret("wrong key"))
		assert.Error(t, "cannot decrypt token", err)
		parsed, err = pagetoken.Decrypt[pagetoken.Cursor[int]](encrypted, pagetoken.Secret("wrong key"))
		assert.Error(t, "cannot decrypt token", err)
	})

	t.Run("wrong type parameter", func(t *testing.T) {
		assert.Error(t,
			"gob: unknown type id or corrupted data",
			must.Get1(pagetoken.ParseOffset(pagetoken.Cursor[int]{
				ReadTimestamp: must.Ok(time.Parse(time.DateTime, "2020-01-01 10:20:30")),
				After:         123,
			}.Encode())),
		)
	})
}

func BenchmarkToken(b *testing.B) {
	b.Run("encode", func(b *testing.B) {
		token := Token{}
		token.After.ID = 123
		b.ResetTimer()
		for b.Loop() {
			token.Encode()
		}
	})

	b.Run("parse", func(b *testing.B) {
		token := Token{}
		token.After.ID = 123
		s := token.Encode()
		b.ResetTimer()
		for b.Loop() {
			_, _ = pagetoken.ParseCursor[KeySet](s)
		}
	})

	b.Run("encrypt", func(b *testing.B) {
		token := Token{}
		b.ResetTimer()
		for b.Loop() {
			token.Encrypt(pagetoken.Secret("secret"))
		}
	})

	b.Run("decrypt", func(b *testing.B) {
		token := Token{}
		encrypted, _ := token.Encrypt(pagetoken.Secret("secret"))
		b.ResetTimer()
		for b.Loop() {
			pagetoken.Decrypt[Token](encrypted, pagetoken.Secret("secret"))
		}
	})

	b.Run("parse decrypt", func(b *testing.B) {
		token := Token{}
		encrypted, _ := token.Encrypt(pagetoken.Secret("secret"))
		b.ResetTimer()
		for b.Loop() {
			_, err := pagetoken.Parse[Token](encrypted, pagetoken.Secret("secret"))
			assert.NoError(b, err)
		}
	})
}
