package slogerr

import (
	"errors"
	"log/slog"

	"go.chrisrx.dev/x/xerrors"
)

// StructuredError is an interface representing an error that contains slog
// attributes.
//
// This is useful for passing around context by wrapping and returning errors,
// allowing other implementations to optionally extract the stored attributes.
type StructuredError interface {
	error
	Attrs() []any
}

// serror implements StructuredError. It stores attributes as a slice of any so
// that it can be passed back to slog as either slog.Attr or pairs of key/value
// args.
type serror struct {
	err   error
	attrs []any
}

// New constructs a new serror implementing StructuredError.
func New(msg string, args ...any) StructuredError {
	return &serror{
		err:   errors.New(msg),
		attrs: args,
	}
}

func (e *serror) Error() string { return e.err.Error() }
func (e *serror) Unwrap() error { return e.err }

// Attrs returns the attributes stored in this error. The error is included
// with all other attributes as a slog.Attr.
func (e *serror) Attrs() []any {
	return append([]any{slog.Any("error", e.err)}, e.attrs...)
}

func (e *serror) LogValue() slog.Value {
	return slog.Group("", e.attrs...).Value
}

// Wrap returns a new StructuredError wrapping the provided error with an
// implementation of StructuredError containing any additional attributes.
// Attributes can be key/value pairs or slog.Attr. If the provided error is
// already a StructuredError, the attributes are appended to the existing
// StructuredError.
//
// If the provided error is nil, then nil is returned. This enables using Wrap
// with a return value without having to check error first.
func Wrap(err error, args ...any) StructuredError {
	if err == nil {
		return nil
	}
	if serr, ok := xerrors.As[*serror](err); ok {
		serr.attrs = append(serr.attrs, args...)
		return serr
	}
	return &serror{
		err:   err,
		attrs: args,
	}
}
