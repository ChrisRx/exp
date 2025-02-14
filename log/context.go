package log

import (
	"context"
	"log/slog"
)

type contextKey struct{}

// Context returns a copy of the parent context with the provided logger stored
// as a value.
func Context(parent context.Context, l *slog.Logger) context.Context {
	return context.WithValue(parent, contextKey{}, l)
}

// FromContext returns a structured logger loaded from the provided context. If
// no logger has been stored, the default logger is returned from the slog
// package.
func FromContext(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && v != nil {
		return v
	}
	return slog.Default()
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

// Discard
func Discard(ctx context.Context) context.Context {
	return Context(ctx, slog.New(&DiscardLogger{}))
}
