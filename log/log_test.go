package log_test

import (
	"log/slog"
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
}
