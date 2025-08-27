package log

import (
	"log/slog"

	"go.chrisrx.dev/x/context"
)

// Key is a [context.Key] for storing a [*slog.Logger] in a [context.Context].
var Key = context.Key[*slog.Logger]()

// Context returns a [context.Context] deferred from the provided parent
// [context.Context], storing a [*slog.Logger].
func Context(parent context.Context, l *slog.Logger) context.Context {
	return Key.WithValue(parent, l)
}

// From returns the [*slog.Logger] stored in the provided [context.Context]. If
// not value is stored, nil is returned.
func From(ctx context.Context) *slog.Logger {
	if !Key.Has(ctx) {
		return slog.Default()
	}
	return Key.Value(ctx)
}
