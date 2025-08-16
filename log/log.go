package log

import (
	"log/slog"
	"os"
	"sync"

	"go.chrisrx.dev/x/env"
)

// Options specifies the configuration for a logger from environment variables.
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

type Option func(*Options)

func WithLevel(lvl slog.Leveler) Option {
	return func(opts *Options) {
		switch lvl := lvl.(type) {
		case slog.Level:
			opts.Level = new(slog.LevelVar)
			opts.Level.Set(lvl)
		case *slog.LevelVar:
			opts.Level = lvl
		}
	}
}

func WithSource(v bool) Option {
	return func(opts *Options) {
		opts.AddSource = v
	}
}

func WithFormat(f Format) Option {
	return func(opts *Options) {
		opts.Format = f
	}
}

func New(opts ...Option) *slog.Logger {
	lv := new(slog.LevelVar)
	lv.Set(slog.LevelInfo)
	o := &Options{
		Level:     lv,
		Format:    TextFormat,
		AddSource: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	hopts := &slog.HandlerOptions{
		AddSource: o.AddSource,
		Level:     o.Level,
	}
	var h slog.Handler
	switch o.Format {
	case JSONFormat:
		h = slog.NewJSONHandler(os.Stdout, hopts)
	default:
		h = slog.NewTextHandler(os.Stdout, hopts)
	}
	return slog.New(&ContextHandler{Handler: h})
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
		var opts Options
		if err := env.Parse(&opts); err != nil {
			slog.Error("cannot parse log options from environment", slog.Any("error", err))
			return
		}
		defaultLevel.Set(opts.Level.Level())
		slog.SetDefault(New(
			WithLevel(defaultLevel),
			WithFormat(opts.Format),
			WithSource(opts.AddSource),
		))
		slog.Debug("default slog configured from environment", slog.Any("options", opts))
	})
}
