//go:generate go tool aliaspkg -docs=all -ignore=Once,WaitGroup

// Package sync provides synchronization primitives not found in the standard
// library [sync] package. Some types are aliases of [sync] types to allow this
// library to be used as a drop-in with extras.
package sync
