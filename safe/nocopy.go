package safe

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
