package log

import (
	"log/slog"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
)

// Options specifies the configuration for a logger from environment variables.
type Options struct {
	Level     slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
	Format    Format     `env:"LOG_FORMAT" envDefault:"text"`
	AddSource bool       `env:"LOG_ADD_SOURCE" envDefault:"true"`
}

func (o Options) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", o.Level.String()),
		slog.Bool("addSource", o.AddSource),
		slog.String("format", o.Format.String()),
	)
}

// defaultOnce ensures that the default logger is only initialized the first
// time SetDefault is called.
var defaultOnce sync.Once

// defaultLevel is a level variable used to change the logging level for the
// default logger.
var defaultLevel = new(slog.LevelVar)

func SetDefaultLevel(lvl slog.Level) {
	defaultLevel.Set(lvl)
}

// SetDefault parses configuration from environment variables and constructs a
// new default slog.Logger. This should be called in the application main
// package so that any packages can easily access a configured slog.Logger.
func SetDefault() {
	defaultOnce.Do(func() {
		opts, err := env.ParseAs[Options]()
		if err != nil {
			slog.Error("cannot parse log options from environment", slog.Any("error", err))
			return
		}
		defaultLevel.Set(opts.Level)
		hopts := &slog.HandlerOptions{
			AddSource: opts.AddSource,
			Level:     defaultLevel,
		}
		var h slog.Handler
		switch opts.Format {
		case JSONFormat:
			h = slog.NewJSONHandler(os.Stdout, hopts)
		default:
			h = slog.NewTextHandler(os.Stdout, hopts)
		}
		h = &ContextHandler{Handler: h}
		slog.SetDefault(slog.New(h))
		slog.Debug("default slog configured from environment", slog.Any("options", opts))
	})
}
