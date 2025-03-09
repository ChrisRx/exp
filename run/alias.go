package run

import (
	"time"

	"github.com/cenkalti/backoff"
)

type BackOff = backoff.BackOff

type ConstantBackOff = backoff.ConstantBackOff

func NewConstantBackOff(d time.Duration) *ConstantBackOff {
	return backoff.NewConstantBackOff(d)
}

type ExponentialBackOff = backoff.ExponentialBackOff

func NewExponetialBackOff() *ExponentialBackOff {
	return backoff.NewExponentialBackOff()
}
