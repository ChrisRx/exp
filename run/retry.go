package run

import (
	"cmp"
	"context"
	"fmt"
	"iter"
	"time"

	"go.chrisrx.dev/x/backoff"
)

type RetryOptions struct {
	InitialInterval     time.Duration
	MaxAttempts         int
	MaxAttemptTime      time.Duration
	MaxElapsedTime      time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64

	b backoff.Backoff
}

func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		InitialInterval: backoff.DefaultMinInterval,
		Multiplier:      backoff.DefaultMultiplier,
		MaxInterval:     backoff.DefaultMaxInterval,
		MaxAttempts:     5,
	}
}

func (ro *RetryOptions) Backoff() backoff.Backoff {
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

func (ro *RetryOptions) Reset() {
	ro.b.Reset()
}

// Retry runs a function periodically based on the provided [Options].
func Retry(ctx context.Context, fn func() error, ro RetryOptions) RetryIterator {
	var attempts int
	return func(yield func(attempts int, err error) bool) {
		if ro.MaxElapsedTime != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, ro.MaxElapsedTime)
			defer cancel()
		}
		ticker := backoff.NewTicker(ro.Backoff())
		defer ticker.Stop()

		for {
			select {
			case <-ticker.Next():
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

type RetryIterator iter.Seq2[int, error]

// Wait ranges over the iterator until no more elements are produced. The last
// error is ignored.
func (r RetryIterator) Wait() {
	_ = r.WaitE()
}

// Wait ranges over the iterator until no more elements are produced. If an
// error is encountered, the last error received will be returned.
func (r RetryIterator) WaitE() error {
	var lastErr error
	for _, err := range r {
		lastErr = err
	}
	return lastErr
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
