package chans

import (
	"go.chrisrx.dev/x/safe"
)

func Collect[T any](ch <-chan T) (s []T) {
	for elem := range ch {
		s = append(s, elem)
	}
	return
}

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
