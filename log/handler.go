package log

import (
	"log/slog"

	"go.chrisrx.dev/x/context"
)

// DiscardLogger is an implementation of [slog.Handler] that discards all
// messages.
type DiscardLogger struct {
	slog.Handler
}

func (h *DiscardLogger) Enabled(ctx context.Context, lvl slog.Level) bool {
	return false
}

// Discard stores a [slog.Logger] in the provided context that discards all log
// messages.
func Discard(ctx context.Context) context.Context {
	return Key.WithValue(ctx, slog.New(&DiscardLogger{}))
}
