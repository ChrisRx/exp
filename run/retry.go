package run

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/samber/lo"
)

const (
	DefaultInitialInterval     = backoff.DefaultInitialInterval
	DefaultMaxAttempts         = 10
	DefaultMaxInterval         = backoff.DefaultMaxInterval
	DefaultMultiplier          = backoff.DefaultMultiplier
	DefaultRandomizationFactor = backoff.DefaultRandomizationFactor
)

type RetryOptions struct {
	InitialInterval     time.Duration
	MaxAttempts         int
	MaxElapsedTime      time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64
}

func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxAttempts:         DefaultMaxAttempts,
	}
}

func (ro RetryOptions) newBackOff() BackOff {
	switch {
	case ro.InitialInterval != 0 && ro.MaxInterval == 0:
		return NewConstantBackOff(ro.InitialInterval)
	case ro.InitialInterval != 0 && ro.MaxInterval != 0 && ro.InitialInterval == ro.MaxInterval:
		return NewConstantBackOff(ro.InitialInterval)
	default:
		return &ExponentialBackOff{
			InitialInterval:     lo.CoalesceOrEmpty(ro.InitialInterval, DefaultInitialInterval),
			RandomizationFactor: lo.CoalesceOrEmpty(ro.RandomizationFactor, DefaultRandomizationFactor),
			Multiplier:          lo.CoalesceOrEmpty(ro.Multiplier, DefaultMultiplier),
			MaxInterval:         lo.CoalesceOrEmpty(ro.MaxInterval, DefaultMaxInterval),
			MaxElapsedTime:      ro.MaxElapsedTime,
			Clock:               backoff.SystemClock,
		}
	}
}

type RetryIterator iter.Seq2[int, error]

func (r RetryIterator) Wait() {
	for range r {
	}
}

func (r RetryIterator) Range(fn func(int, error)) {
	_ = r.RangeE(func(attempts int, err error) error {
		fn(attempts, err)
		return nil
	})
}

func (r RetryIterator) RangeE(fn func(int, error) error) error {
	for attempts, err := range r {
		if err := fn(attempts, err); err != nil {
			return err
		}
	}
	return nil
}

func Retry(ctx context.Context, fn func() error, ro RetryOptions) RetryIterator {
	var attempts int
	return func(yield func(attempts int, err error) bool) {
		ticker := backoff.NewTicker(ro.newBackOff())
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				attempts++

				if err := func() (reterr error) {
					defer func() {
						if r := recover(); r != nil {
							switch t := r.(type) {
							case error:
								reterr = fmt.Errorf("panic: %w", t)
							default:
								reterr = fmt.Errorf("panic: %v", t)
							}
						}
					}()

					return fn()
				}(); err != nil {
					if !yield(attempts, err) {
						return
					}
					if ro.MaxAttempts != 0 && attempts >= ro.MaxAttempts {
						return
					}
					continue
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}
}
