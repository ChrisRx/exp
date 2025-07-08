package sync

import "sync"

type (
	Locker    = sync.Locker
	Map       = sync.Map
	Mutex     = sync.Mutex
	Once      = sync.Once
	Pool      = sync.Pool
	RWMutex   = sync.RWMutex
	WaitGroup = sync.WaitGroup
)

func OnceFunc(f func()) func() {
	return sync.OnceFunc(f)
}

func OnceValue[T any](f func() T) func() T {
	return sync.OnceValue(f)
}

func OnceValues[T1, T2 any](f func() (T1, T2)) func() (T1, T2) {
	return sync.OnceValues(f)
}
