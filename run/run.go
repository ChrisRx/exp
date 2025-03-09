package run

import (
	"context"
	"errors"
	"time"
)

func Every[T Interval](ctx context.Context, fn func(), interval T) {
	Retry(ctx, asRetryFunc(fn), newRetryOptions(interval)).Wait()
}

func Until[T Interval](ctx context.Context, fn func() bool, interval T) {
	Retry(ctx, asRetryFunc(fn), newRetryOptions(interval)).Wait()
}

func UntilE[T Interval](ctx context.Context, fn func() error, interval T) {
	Retry(ctx, asRetryFunc(fn), newRetryOptions(interval)).Wait()
}

type Interval interface {
	time.Duration | RetryOptions
}

func newRetryOptions[T Interval](interval T) RetryOptions {
	switch t := any(interval).(type) {
	case time.Duration:
		return RetryOptions{
			InitialInterval: t,
		}
	case RetryOptions:
		return t
	default:
		panic("unreachable")
	}
}

var errContinue = errors.New("continue")

func asRetryFunc[T interface {
	func() | func() bool | func() error
}](fn T) func() error {
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
