package log

import (
	"context"
	"log/slog"
)

// TODO(chrism): need handler that will deduplicate messages in certain
// situations.

type attrContextKey struct{}

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(attrContextKey{}).([]any); ok {
		r.Add(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

type DiscardLogger struct {
	slog.Handler
}

func (h *DiscardLogger) Enabled(ctx context.Context, lvl slog.Level) bool {
	return false
}
