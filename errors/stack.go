package errors

import (
	"strings"

	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/stack"
)

type Frames []stack.Frame

func (f Frames) String() string {
	return strings.Join(slices.Map(f, func(f stack.Frame) string {
		return f.String()
	}), "\n    ")
}

type StackError interface {
	error

	// Trace returns a stack trace for this error.
	Trace() Frames

	isStackError()
}

type stackError struct {
	error

	frames []stack.Frame
}

var _ StackError = (*stackError)(nil)

func Stack(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(StackError); ok {
		return err
	}
	return &stackError{
		error:  err,
		frames: stack.Trace(1),
	}
}

func (e *stackError) Error() string { return e.error.Error() }
func (e *stackError) Unwrap() error { return e.error }

func (e *stackError) Trace() Frames {
	return e.frames
}

func (e *stackError) isStackError() {}
