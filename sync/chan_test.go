package sync_test

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/sync"
)

func TestChan(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		type S struct {
			ch sync.Chan[int]
		}

		var s S
		s.ch.New(1)
		s.ch.Load() <- 10
		assert.Equal(t, 10, <-s.ch.Load())
	})

	t.Run("send unbuffered", func(t *testing.T) {
		messages := make([]int, 10)
		for i := range 10 {
			messages[i] = i
		}

		ch := sync.NewChan[int](0)

		go func() {
			defer ch.Close()

			ch := ch.Load()
			for _, v := range messages {
				ch <- v
			}
		}()

		assert.Equal(t,
			messages,
			slices.FromChan(ch.Load()),
		)
	})

	t.Run("reset", func(t *testing.T) {
		expected := "successful message"
		ch := sync.NewChan[string](1)
		ch.Load() <- "this message will be dropped"
		ch.Reset() <- expected

		assert.Equal(t, expected, <-ch.Recv())
	})

	t.Run("size", func(t *testing.T) {
		var ch sync.Chan[int]
		assert.Equal(t, 8, int(unsafe.Sizeof(ch)))
	})

	t.Run("capacity", func(t *testing.T) {
		ch := sync.NewChan[int](0)

		assert.Equal(t, 0, ch.Cap(), "initial capacity")
		ch.New(5)
		assert.Equal(t, 5, ch.Cap(), "set capacity to 5")
		ch.Reset()
		assert.Equal(t, 5, ch.Cap(), "same capacity after reset")
	})

	t.Run("load after close", func(t *testing.T) {
		ch := sync.NewChan[int](0)

		go ch.Send(10)

		assert.Equal(t, 10, <-ch.Recv())
		ch.Close()
		assert.Equal(t, true, ch.Closed())
		assert.PanicsWithError(t, "send on closed channel", func() {
			ch.Load() <- 10
		})
	})

	t.Run("send timeout", func(t *testing.T) {
		var (
			n   = 5
			buf = 2
		)

		messages := make([]int, n)
		for i := range n {
			messages[i] = i
		}
		ch := sync.NewChan[int](buf)
		assert.Equal(t, false, ch.TrySend(messages...))
		assert.Equal(t,
			messages[:buf],
			slices.FromChan(ch.CloseAndRecv()),
			"send multiple values timeout",
		)
	})

	t.Run("send on closed", func(t *testing.T) {
		ch := sync.NewChan[int](0)
		close(ch.Load())
		assert.NotPanics(t, func() {
			ch.Send(20)
		})
	})
}
