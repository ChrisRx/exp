package run

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"go.chrisrx.dev/x/backoff"
	"go.chrisrx.dev/x/must"
)

type Options struct {
	InitialInterval     time.Duration
	MaxAttempts         int
	MaxAttemptTime      time.Duration
	MaxElapsedTime      time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64

	b backoff.Backoff
}

func DefaultOptions() Options {
	return Options{
		InitialInterval: backoff.DefaultMinInterval,
		Multiplier:      backoff.DefaultMultiplier,
		MaxInterval:     backoff.DefaultMaxInterval,
		MaxAttempts:     5,
	}
}

func (ro *Options) Backoff() backoff.Backoff {
	switch {
	case ro.InitialInterval != 0 && ro.MaxInterval == 0:
		ro.b.MinInterval = ro.InitialInterval
		ro.b.MaxInterval = ro.InitialInterval
	case ro.InitialInterval != 0 && ro.MaxInterval != 0 && ro.InitialInterval == ro.MaxInterval:
		ro.b.MinInterval = ro.InitialInterval
		ro.b.MaxInterval = ro.InitialInterval
	default:
		ro.b.MinInterval = cmp.Or(ro.InitialInterval, backoff.DefaultMinInterval)
		ro.b.MaxInterval = cmp.Or(ro.MaxInterval, backoff.DefaultMaxInterval)
		ro.b.Multiplier = cmp.Or(ro.Multiplier, backoff.DefaultMultiplier)
	}
	return ro.b
}

func (ro *Options) Reset() {
	ro.b.Reset()
}

// Every runs a function periodically for the provided interval. It runs
// indefinitely or until the context is done.
//
// The interval can be either a [time.Duration] or, if more complex retry logic
// is required, a [Options].
func Every[T Interval](ctx context.Context, fn func(), interval T) {
	ro := retryOptionsFromInterval(interval)
	// ignore these user provided values so this runs indefinitely
	ro.MaxAttempts = 0
	ro.MaxElapsedTime = 0
	_ = Do(ctx, func(ctx context.Context) (bool, error) {
		fn()
		return false, nil
	}, ro)
}

// Until runs a function periodically for the provided interval. This is used
// for running logic until something is successful or until the context is
// done.
//
// Until has different behaviors depending on the retry function that is passed
// in. If the function returns an error, it will run until no error is
// encountered. If given a retry function returning a bool, then it will run
// until true is returned.
//
// The interval can be either a [time.Duration] or, if more complex retry logic
// is required, a [Options].
func Until[R RetryFunc, T Interval](ctx context.Context, fn R, interval T) error {
	return Do(ctx, func(ctx context.Context) (bool, error) {
		ok, err := asRetryFunc(fn)()
		return ok, err
	}, retryOptionsFromInterval(interval))
}

// Unless runs a function periodically for the provided interval. This is used
// for running logic until something is unsuccessful. It is the inverse of
// [Until].
//
// Unless has different behaviors depending on the retry function that is passed
// in. If the function returns an error, it will run until an error is
// encountered. If given a retry function returning a bool, then it will run
// until false is returned.
//
// The interval can be either a [time.Duration] or, if more complex retry logic
// is required, a [Options].
func Unless[R RetryFunc, T Interval](ctx context.Context, fn R, interval T) error {
	return Do(ctx, func(ctx context.Context) (bool, error) {
		ok, err := asRetryFunc(fn)()
		return !ok, err
	}, retryOptionsFromInterval(interval))
}

func Do(parent context.Context, fn func(context.Context) (bool, error), ro Options) error {
	if ro.MaxElapsedTime != 0 {
		var cancel context.CancelFunc
		parent, cancel = context.WithTimeout(parent, ro.MaxElapsedTime)
		defer cancel()
	}
	ticker := backoff.NewTicker(ro.Backoff())
	defer ticker.Stop()

	var attempts int
	for {
		select {
		case <-ticker.Next():
			attempts++
			ctx := parent
			if ro.MaxAttemptTime != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(parent, ro.MaxAttemptTime)
				defer cancel()
			}
			done, err := func() (_ bool, reterr error) {
				defer must.Catch(&reterr)
				return fn(ctx)
			}()
			if done {
				return err
			}
			if ro.MaxAttempts != 0 && attempts >= ro.MaxAttempts {
				if err != nil {
					return fmt.Errorf("max attempts: %w", err)
				}
				return fmt.Errorf("max attempts: %d", attempts)
			}
		case <-parent.Done():
			return parent.Err()
		}
	}
}

type Interval interface {
	time.Duration | Options
}

func retryOptionsFromInterval[T Interval](interval T) Options {
	switch t := any(interval).(type) {
	case time.Duration:
		return Options{
			InitialInterval: t,
		}
	case Options:
		return t
	default:
		panic("unreachable")
	}
}

type RetryFunc interface {
	func() bool | func() error | func() (bool, error)
}

func asRetryFunc[R RetryFunc](fn R) func() (bool, error) {
	switch fn := any(fn).(type) {
	case func() bool:
		return func() (bool, error) {
			return fn(), nil
		}
	case func() error:
		return func() (bool, error) {
			err := fn()
			return err == nil, err
		}
	case func() (bool, error):
		return fn
	default:
		panic("unreachable")
	}
}
