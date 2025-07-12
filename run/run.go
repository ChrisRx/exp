package run

import (
	"context"
	"errors"
	"time"
)

type Interval interface {
	time.Duration | *RetryOptions
}

type RetryFunc interface {
	func() | func() bool | func() error
}

// hack to avoid errcheck
type error_ = error

// Every runs a function periodically for the provided interval.
//
// The interval can be either a [time.Duration] or, if more complex retry logic
// is required, a [RetryOptions]. Given a [time.Duration], this will run the
// function forever at a constant interval.
//
// The context passed in can be used to return early, regardless of the
// interval provided. If Every returns early, the last error (if any) will be
// returned.
func Every[R RetryFunc, T Interval](ctx context.Context, fn R, interval T) error_ {
	return Retry(ctx, asRetryFunc(fn), newRetryOptions(interval)).WaitE()
}

// Until runs a function periodically for the provided interval. This is used
// for running logic until something is successful. Until has different
// behaviors depending on the retry function that is passed in. If the function
// returns an error, it will run until no error is returned. If given a retry
// function returning a bool, then it will run until true is returned.
//
// The interval can be either a [time.Duration] or, if more complex retry logic
// is required, a [RetryOptions]. Given a [time.Duration], this will run the
// function forever at a constant interval.
//
// The context passed in can be used to return early, regardless of the
// interval provided. If Until returns early, the last error (if any) will be
// returned.
func Until[R RetryFunc, T Interval](ctx context.Context, fn R, interval T) error_ {
	return Retry(ctx, asRetryFunc(fn), newRetryOptions(interval)).WaitE()
}

func newRetryOptions[T Interval](interval T) *RetryOptions {
	switch t := any(interval).(type) {
	case time.Duration:
		return &RetryOptions{
			InitialInterval: t,
		}
	case RetryOptions:
		return &t
	default:
		panic("unreachable")
	}
}

var errContinue = errors.New("continue")

func asRetryFunc[T RetryFunc](fn T) func() error {
	switch fn := any(fn).(type) {
	case func():
		return func() error {
			fn()
			return errContinue
		}
	case func() bool:
		return func() error {
			if fn() {
				return nil
			}
			return errContinue
		}
	case func() error:
		return fn
	default:
		panic("unreachable")
	}
}
