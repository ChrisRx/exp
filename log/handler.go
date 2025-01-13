package log

import (
	"context"
	"log/slog"
)

type attrContextKey struct{}

type ContextHandler struct {
	slog.Handler

	lvl slog.LevelVar
}

func (h *ContextHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	if lvl < h.lvl.Level() {
		return false
	}

	return h.Handler.Enabled(ctx, lvl)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(attrContextKey{}).([]any); ok {
		r.Add(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a copy of the provided context adding log attributes as
// context values.
//
// This allows slog.Handler implementations, such as ContextHandler, to extract
// attributes automatically when using slog functions that accept a
// context.Context.
func WithAttrs(ctx context.Context, args ...any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if v, ok := ctx.Value(attrContextKey{}).([]any); ok {
		v = append(v, args...)
		return context.WithValue(ctx, attrContextKey{}, v)
	}
	v := make([]any, 0)
	v = append(v, args...)
	return context.WithValue(ctx, attrContextKey{}, v)
}
