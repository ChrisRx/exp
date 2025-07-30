package safe

// NoCompare is a type that can be added to the first field of a struct to
// prevent that struct from being comparable.
type NoCompare [0]func()

// NoCopy implements [sync.Locker] which means that adding it to a struct
// triggers the go vet copylocks check. This makes it useful to add to structs
// that shouldn't be copied after first use:
//
//	type S struct {
//		_ safe.NoCopy
//	}
//
// It should not be embedded, since it would expose the [sync.Locker] methods.
//
// https://golang.org/issues/8005#issuecomment-190753527
type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}

// NoTypeConversion is a type that can be added as a field to a struct to
// prevent unsafe type conversion.
type NoTypeConversion[T any] [0]*T
