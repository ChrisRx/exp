package log

import (
	"fmt"
	"log/slog"
	"strings"
)

type Format int

const (
	// TextFormat sets log output to logfmt.
	TextFormat Format = iota
	// JSONFormat sets log output format to JSON.
	JSONFormat
)

func (f Format) String() string {
	switch f {
	case TextFormat:
		return "text"
	case JSONFormat:
		return "json"
	default:
		return "text"
	}
}

// MarshalText implements encoding.TextMarshaler by calling Format.String.
func (f Format) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It accepts any string
// produced by Format.MarshalText, ignoring case.
func (f *Format) UnmarshalText(data []byte) error {
	switch strings.ToLower(string(data)) {
	case "json":
		*f = JSONFormat
	case "text":
		*f = TextFormat
	default:
		return fmt.Errorf("received invalid Format: %q", data)
	}
	return nil
}

type Options struct {
	Level     *slog.LevelVar `env:"LOG_LEVEL" default:"INFO"`
	Format    Format         `env:"LOG_FORMAT" default:"text"`
	AddSource bool           `env:"LOG_ADD_SOURCE" default:"true"`
}

func (o Options) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", o.Level.Level().String()),
		slog.Bool("addSource", o.AddSource),
		slog.String("format", o.Format.String()),
	)
}
