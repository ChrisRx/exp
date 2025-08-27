package log_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/log"
)

func newLevelVar(lvl slog.Level) *slog.LevelVar {
	lv := new(slog.LevelVar)
	lv.Set(lvl)
	return lv
}

func TestLog(t *testing.T) {
	assert.WithEnviron(t, map[string]string{
		"LOG_LEVEL":      "DEBUG",
		"LOG_FORMAT":     "json",
		"LOG_ADD_SOURCE": "true",
	}, func() {
		var opts log.Options
		assert.NoError(t, env.Parse(&opts))
		assert.Equal(t, log.Options{
			Level:     newLevelVar(slog.LevelDebug),
			Format:    log.JSONFormat,
			AddSource: true,
		}, opts)
	})

	t.Run("context", func(t *testing.T) {
		ctx := t.Context()
		l := log.New()
		ctx = log.Context(ctx, l)
		assert.Equal(t, l, log.Key.Value(ctx))
	})

	t.Run("discard", func(t *testing.T) {
		ctx := t.Context()

		var buf bytes.Buffer
		ctx = log.Context(ctx, log.New(
			log.WithSource(false),
			log.WithOutput(&buf),
			log.WithRemoveAttrs(slog.LevelKey, slog.TimeKey),
		))

		assert.Equal(t, "msg=testing", func() string {
			defer buf.Reset()
			log.From(ctx).Info("testing")
			return strings.TrimSuffix(buf.String(), "\n")
		}())

		ctx = log.Discard(ctx)
		assert.Equal(t, "", func() string {
			defer buf.Reset()
			log.Key.Value(ctx).Info("testing")
			return strings.TrimSuffix(buf.String(), "\n")
		}())
	})
}
