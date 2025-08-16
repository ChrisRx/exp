package sync

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

func TestSemaphore(t *testing.T) {
	t.Run("", func(t *testing.T) {
		sema := NewSemaphore(5)

		sema.Acquire(5)

		go func() {
			for range 5 {
				time.Sleep(25 * time.Millisecond)
				sema.Release()
			}
		}()

		sema.Acquire(5)
		sema.ReleaseN(5)
		var empty bool
		select {
		case <-sema.ch.Recv():
		default:
			empty = true
		}
		assert.Equal(t, true, empty, "semaphore channel is empty")
	})

	t.Run("zero value", func(t *testing.T) {
		var sema Semaphore
		sema.Acquire(100)
	})

	t.Run("weight greater than capacity", func(t *testing.T) {
		sema := NewSemaphore(5)
		sema.Acquire(6)
	})
}
