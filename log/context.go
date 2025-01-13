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
