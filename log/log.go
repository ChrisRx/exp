//go:generate go tool aliaspkg -docs=decls -include Panic,Panicf,Panicln,Print,Printf,Println

package log

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/strings"
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

// Fatal is similar to [log.Fatal], but uses the default [slog.Logger]. It is
// meant to be used at the end of an error chain only.
//
// If GO_BACKTRACE=1 environment variable is set, any arguments containing a
// stack trace will be printed directly to stderr. Only the first argument
// containing a stack trace will be printed.
func Fatal(v ...any) {
	msg := fmt.Sprint(v...)
	if isBacktraceEnabled() {
		printBacktrace(msg, v)
	}
	slog.Error(msg)
	os.Exit(1)
}

// Fatalf is similar to [log.Fatalf], but uses the default [slog.Logger]. It is
// meant to be used at the end of an error chain only.
//
// If GO_BACKTRACE=1 environment variable is set, any arguments containing a
// stack trace will be printed directly to stderr. Only the first argument
// containing a stack trace will be printed.
func Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if isBacktraceEnabled() {
		printBacktrace(msg, v)
	}
	slog.Error(msg)
	os.Exit(1)
}

// Fatalln is similar to [log.Fatalln], but uses the default [slog.Logger]. It
// is meant to be used at the end of an error chain only.
//
// If GO_BACKTRACE=1 environment variable is set, any arguments containing a
// stack trace will be printed directly to stderr. Only the first argument
// containing a stack trace will be printed.
func Fatalln(v ...any) {
	msg := fmt.Sprintln(v...)
	if isBacktraceEnabled() {
		printBacktrace(msg, v)
	}
	slog.Error(msg)
	os.Exit(1)
}

const backtraceEnvVar = "GO_BACKTRACE"

func isBacktraceEnabled() bool {
	if value, ok := os.LookupEnv(backtraceEnvVar); ok && value == "1" {
		return true
	}
	return false
}

// checkBacktrace checks if any arguments implement [errors.StackError] and
// prints the full stack trace to stderr. Arguments are checked in order and if
// multiple stack traces are provided, only the first is printed.
//
// This behavior is opt-in, requiring environment variable GO_BACKTRACE=1 to be
// set.
func printBacktrace(msg string, args []any) {
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			if err, ok := errors.As[errors.StackError](err); ok {
				if v, ok := os.LookupEnv(backtraceEnvVar); ok && v == "1" {
					slog.Error(msg)
					fmt.Fprintln(os.Stderr, strings.Join(slices.Map(err.Trace(), strings.ToString), "\n\t"))
					os.Exit(1)
				}
			}
		}
	}
}
