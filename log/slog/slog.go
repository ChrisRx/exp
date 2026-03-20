package slog

import (
	"fmt"
	"log/slog"
)

//go:generate go tool aliaspkg -docs=all log/slog

// Stringf returns an Attr for a formatted string value.
func Stringf(key, format string, a ...any) slog.Attr {
	return slog.Attr{Key: key, Value: slog.StringValue(fmt.Sprintf(format, a...))}
}
