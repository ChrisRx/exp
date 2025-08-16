// Package chans defines various functions for working with Go channels.
package chans

import (
	"go.chrisrx.dev/x/safe"
)

// Collect receives from a channel until the channel is closed and returns a
// slice of the received values. There is no way provided for cancellation so
// care should be used when using this function.
func Collect[T any](ch <-chan T) (s []T) {
	for elem := range ch {
		s = append(s, elem)
	}
	return
}

// CollectN receives from a channel until the channel is closed or it receives
// n values. It returns a slice of the received values.
func CollectN[T any](ch <-chan T, n int) (s []T) {
	var i int
	for elem := range ch {
		s = append(s, elem)
		i++
		if i >= n {
			break
		}
	}
	return
}

// Drain closes the provided channel and receives all the remaining values. Any
// values received are sent to a temporary channel that is open until the
// remaining values are received.
func Drain[T any](ch chan T) <-chan T {
	safe.Close(ch)
	out := make(chan T)
	go func() {
		defer close(out)
		for elem := range ch {
			out <- elem
		}
	}()
	return out
}
