package sync

import "sync"

type (
	// Cond is a type alias of [sync.Cond].
	Cond = sync.Cond
	// Locker is a type alias of [sync.Locker].
	Locker = sync.Locker
	// Map is a type alias of [sync.Map].
	Map = sync.Map
	// Mutex is a type alias of [sync.Mutex].
	Mutex = sync.Mutex
	// Once is a type alias of [sync.Once].
	Once = sync.Once
	// Pool is a type alias of [sync.Pool].
	Pool = sync.Pool
	// RWMutex is a type alias of [sync.RWMutex].
	RWMutex = sync.RWMutex
	// WaitGroup is a type alias of [sync.WaitGroup].
	WaitGroup = sync.WaitGroup
)
