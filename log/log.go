//go:generate go tool aliaspkg -docs=decls -include Fatal,Fatalf,Fatalln,Panic,Panicf,Panicln,Print,Printf,Println

package log

import (
	"log/slog"
	"sync"

	"go.chrisrx.dev/x/env"
)

// New constructs a new [*slog.Logger] with the provided options.
func New(opts ...Option) *slog.Logger {
	return slog.New(NewOptions(opts...).New())
}

// DefaultLevel is a level variable used to change the logging level for the
// default logger.
var DefaultLevel = new(slog.LevelVar)

// defaultOnce ensures that the default logger is only initialized the first
// time SetDefault is called.
var defaultOnce sync.Once

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
		DefaultLevel.Set(opts.Level.Level())
		slog.SetDefault(New(WithOptions(opts)))
		slog.Debug("default slog configured from environment", slog.Any("options", opts))
	})
}
