// Package backoff provides a simple implementation of exponential backoff.
package backoff

import (
	"math/rand/v2"
	"time"
)

// Default values for [Backoff].
const (
	DefaultMinInterval = 100 * time.Millisecond
	DefaultMaxInterval = 60 * time.Second
	DefaultMultiplier  = 1.5
)

// Backoff is a simple exponential backoff implementation. The zero Backoff is
// valid and will use default configuration values. It is not thread-safe.
type Backoff struct {
	// MinInterval is the initial interval for the backoff. It is used to set the
	// lower bound on the range of intervals returned from [Backoff.Next].
	MinInterval time.Duration
	// MaxInterval is the maximum interval that should be returned from
	// [Backoff.Next].
	MaxInterval time.Duration
	// Multiplier is the Multiplier used when computing the next interval.
	Multiplier float64
	// Jitter is used to specify a range of randomness for intervals.
	Jitter time.Duration

	cur time.Duration
}

func (b *Backoff) init() {
	if b.Multiplier == 0 {
		b.Multiplier = DefaultMultiplier
	}
	if b.MinInterval == 0 {
		b.MinInterval = DefaultMinInterval
	}
	if b.MaxInterval == 0 {
		b.MaxInterval = DefaultMaxInterval
	}
	if b.MinInterval < 0 || b.MaxInterval < 0 || b.Multiplier < 0 {
		panic("cannot provide negative values for backoff")
	}
}

// Next returns the next interval to wait.
func (b *Backoff) Next() (next time.Duration) {
	b.init()

	defer func() {
		if b.Jitter > 0 {
			next = applyJitter(b.cur, b.Jitter)
		}
	}()

	// When the min/max interval matches, this is effectively a constant
	// interval. Alternatively, if the min interval is less than max interval,
	// there was probably a misconfiguration of the backoff, so we also just
	// return early with the specified min interval.
	if b.MinInterval == b.MaxInterval || b.MinInterval > b.MaxInterval {
		return b.MinInterval
	}

	switch {
	case b.cur == 0:
		// The current interval will be zero the first time Next is called. This
		// lets us ensure the first time Next is called will produce the min
		// interval specified rather than the next interval.
		b.cur = b.MinInterval
		return b.cur
	case b.cur == b.MaxInterval:
		// When the current interval matches the max interval, we can return early
		// since there is no need to adjust the interval any further.
		return b.cur
	case float64(b.cur) >= float64(b.MaxInterval)/b.Multiplier:
		// If the next interval will be greater than the max interval, there is no
		// reason to further update the current interval. By setting it to the max
		// interval will ensure that this short circuits earlier on future calls.
		b.cur = b.MaxInterval
		return b.cur
	}

	// Finally, the current interval can safely be updated with the given
	// configuration.
	b.cur = time.Duration(float64(b.cur) * b.Multiplier)
	return b.cur
}

func applyJitter(d, jitter time.Duration) time.Duration {
	min := d - jitter
	max := d + jitter
	return rand.N(max-min) + min
}

// Reset returns the interval back to the initial state.
func (b *Backoff) Reset() {
	b.cur = 0
}
