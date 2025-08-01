package backoff

import (
	"math/rand/v2"
	"time"
)

const (
	DefaultMinInterval = 100 * time.Millisecond
	DefaultMaxInterval = 60 * time.Second
	DefaultMultiplier  = 1.5
)

type Backoff struct {
	MinInterval time.Duration
	MaxInterval time.Duration
	Multiplier  float64
	Jitter      time.Duration

	cur time.Duration
}

func (b *Backoff) Next() (next time.Duration) {
	defer func() {
		if b.Jitter > 0 {
			min := b.cur - b.Jitter
			max := b.cur + b.Jitter
			next = rand.N(max-min) + min
		}
	}()

	if b.Multiplier == 0 {
		b.Multiplier = DefaultMultiplier
	}
	if b.MinInterval == 0 {
		b.MinInterval = DefaultMinInterval
	}
	if b.MaxInterval == 0 {
		b.MaxInterval = DefaultMaxInterval
	}

	// When the min/max interval matches, this is effectively a constant
	// interval.
	if b.MinInterval == b.MaxInterval {
		return b.MinInterval
	}

	if b.cur == 0 {
		b.cur = b.MinInterval
		return b.cur
	}

	if b.cur == b.MaxInterval {
		return b.cur
	}

	if float64(b.cur) >= float64(b.MaxInterval)/b.Multiplier {
		b.cur = b.MaxInterval
		return b.cur
	}

	b.cur = time.Duration(float64(b.cur) * b.Multiplier)
	return b.cur
}

func (b *Backoff) Reset() {
	b.cur = b.MinInterval
}
