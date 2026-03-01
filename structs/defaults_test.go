package structs_test

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/env/testdata/pg"
	"go.chrisrx.dev/x/structs"
)

func TestDefaults(t *testing.T) {
	t.Run("", func(t *testing.T) {
		cfg := structs.Defaults(pg.Config{})

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
		assert.Equal(t, pg.Prefer, cfg.SSLMode)
	})

	t.Run("struct pointer", func(t *testing.T) {
		var cfg pg.Config
		structs.Defaults(&cfg)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
		assert.Equal(t, pg.Prefer, cfg.SSLMode)

		cfg.ConnectTimeout = 60 * time.Second
		cfg.SSLMode = pg.Require

		structs.Defaults(&cfg)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, 60*time.Second, cfg.ConnectTimeout)
		assert.Equal(t, pg.Require, cfg.SSLMode)
	})

	t.Run("uninitialized", func(t *testing.T) {
		assert.Panic(t, "value provided to Defaults must be initialized", func() {
			var cfg *pg.Config
			structs.Defaults(cfg)
		})
	})

	t.Run("override defaults", func(t *testing.T) {
		cfg := structs.Defaults(pg.Config{
			ConnectTimeout: 60 * time.Second,
		})

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, 60*time.Second, cfg.ConnectTimeout)
		assert.Equal(t, pg.Prefer, cfg.SSLMode)
	})
}

func TestDefaultsFor(t *testing.T) {
	t.Run("", func(t *testing.T) {
		cfg := structs.DefaultsFor[pg.Config]()

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
		assert.Equal(t, pg.Prefer, cfg.SSLMode)
	})
}
