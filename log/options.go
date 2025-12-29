package log

import (
	"cmp"
	"io"
	"log/slog"
	"os"

	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/slices"
)

// Options specifies the configuration for a logger from environment variables.
type Options struct {
	_ env.Deferred `env:"LOG_DEFAULT" method:"SetDefault" default:"true"`

	Level       *slog.LevelVar `env:"LOG_LEVEL" default:"INFO"`
	Format      Format         `env:"LOG_FORMAT" default:"text"`
	AddSource   bool           `env:"LOG_ADD_SOURCE" default:"true"`
	RemoveAttrs []string       `env:"LOG_REMOVE_ATTRS"`

	out io.Writer
}

func (o Options) SetDefault() {
	DefaultLevel.Set(o.Level.Level())
	slog.SetDefault(slog.New(o.New()))
	slog.Debug("default slog configured from environment", slog.Any("options", o))
}

// NewOptions constructs an [Options] with the provided options. Options not
// explicitly set will be initialized with defaults.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Level:     new(slog.LevelVar),
		Format:    TextFormat,
		AddSource: true,
		out:       os.Stdout,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o Options) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", o.Level.Level().String()),
		slog.Bool("addSource", o.AddSource),
		slog.String("format", o.Format.String()),
		slog.Any("removeAttrs", o.RemoveAttrs),
	)
}

// New constructs a new [slog.Logger] with the values of [Options].
func (o Options) New() slog.Handler {
	hopts := &slog.HandlerOptions{
		AddSource: o.AddSource,
		Level:     cmp.Or(o.Level, new(slog.LevelVar)),
	}
	if len(o.RemoveAttrs) > 0 {
		hopts.ReplaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			if slices.ContainsFunc(o.RemoveAttrs, func(key string) bool {
				return key == attr.Key
			}) {
				return slog.Attr{}
			}
			return attr
		}
	}
	if o.out == nil {
		o.out = os.Stdout
	}
	switch o.Format {
	case JSONFormat:
		return slog.NewJSONHandler(o.out, hopts)
	case TextFormat:
		return slog.NewTextHandler(o.out, hopts)
	default:
		return slog.NewTextHandler(o.out, hopts)
	}
}

// NewLogger constructs a new [*slog.Logger] from the current values of
// [Options].
func (o Options) NewLogger() *slog.Logger {
	return New(WithOptions(o))
}

// SetOutput sets the output for any derived loggers.
func (o *Options) SetOutput(w io.Writer) {
	o.out = w
}

type Option func(*Options)

// WithFormat is an option setting the logger format.
func WithFormat(f Format) Option {
	return func(o *Options) {
		o.Format = f
	}
}

// WithLevel is an option for specifying the level of a logger.
func WithLevel(lvl slog.Leveler) Option {
	return func(opts *Options) {
		switch lvl := lvl.(type) {
		case *slog.LevelVar:
			opts.Level = lvl
		default:
			opts.Level = new(slog.LevelVar)
			opts.Level.Set(lvl.Level())
		}
	}
}

// WithOptions allows passing [Options] to completely override the existing
// options.
func WithOptions(opts Options) Option {
	return func(o *Options) {
		*o = opts
	}
}

// WithOutput is an option setting the logger output.
func WithOutput(w io.Writer) Option {
	return func(o *Options) {
		o.out = w
	}
}

// WithRemoveAttrs is an option removing any [slog.Attr] that match the
// provided keys.
func WithRemoveAttrs(attrs ...string) Option {
	return func(o *Options) {
		o.RemoveAttrs = attrs
	}
}

// WithSource is an option specifying if the source attribute is added to
// records.
func WithSource(v bool) Option {
	return func(o *Options) {
		o.AddSource = v
	}
}
