package assert_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

type T struct {
	*testing.T
}

func NoFatal(t *testing.T) *T { return &T{T: t} }

func (t *T) Fatal(args ...any)                 { t.Helper(); t.Log(args...) }
func (t *T) Fatalf(format string, args ...any) { t.Helper(); t.Logf(format, args...) }

func TestAssert(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		assert.Equal(t, "testing", "testing")
		assert.Equal(NoFatal(t), "comparable", "compareable")
		type Backoff struct {
			MinInterval time.Duration
			MaxInterval time.Duration
			Multiplier  float64
			Jitter      time.Duration
			cur         time.Duration
		}
		b1 := &Backoff{}
		b2 := &Backoff{}
		assert.Equal(NoFatal(t), b1, b2)
		b2.MinInterval = 100 * time.Millisecond
		assert.Equal(NoFatal(t), b1, b2)
	})

	t.Run("ElementsMatch", func(t *testing.T) {
		assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, []int{5, 3, 2, 1, 4})
		assert.ElementsMatch(NoFatal(t), []int{1, 2, 3, 4, 5}, []int{5, 3, 2, 1, 6})
	})

	t.Run("Panic", func(t *testing.T) {
		ch := make(chan struct{})
		close(ch)
		assert.Panic(t, "send on closed channel", func() {
			ch <- struct{}{}
		})
		assert.Panic(NoFatal(t), "some other panic", func() {
			ch <- struct{}{}
		})
		assert.Panic(NoFatal(t), nil, func() {})
		assert.Panic(NoFatal(t), "some panic", func() {})
		assert.NoPanic(NoFatal(t), func() {
			panic("shouldn't panic")
		})
	})

	t.Run("Error", func(t *testing.T) {
		err := errors.New("base error")
		err2 := fmt.Errorf("wrapped: %w", err)
		err3 := fmt.Errorf("final: %w", err2)
		assert.Error(t, err3, err3)
		assert.Error(t, err3, err2)
		assert.Error(NoFatal(t), err3, fmt.Errorf("whoops"))

		assert.Error(t, "this is an error", fmt.Errorf("this is an error"))
		assert.Error(NoFatal(t), "this is an error", fmt.Errorf("this is an error extra"))
		assert.NoError(NoFatal(t), fmt.Errorf("this was an error"))
	})

	t.Run("WithinDuration", func(t *testing.T) {
		var now time.Time
		assert.WithinDuration(t, now, now.Add(100*time.Millisecond), 100*time.Millisecond)
		assert.WithinDuration(NoFatal(t), now, now.Add(101*time.Millisecond), 100*time.Millisecond)
	})

	t.Run("Between", func(t *testing.T) {
		var now time.Time
		assert.Between(t, now.Add(-time.Hour), now.Add(time.Hour), now.Add(1*time.Hour))
		assert.Between(NoFatal(t), now.Add(-time.Hour), now.Add(time.Hour), now.Add(2*time.Hour))
		assert.Between(NoFatal(t), 5, 10, 5)
		assert.Between(NoFatal(t), 5, 10, 8)
		assert.Between(NoFatal(t), 5, 10, 10)
		assert.Between(NoFatal(t), 5, 10, 4)
		assert.Between(NoFatal(t), 5, 10, 11)
	})
}
