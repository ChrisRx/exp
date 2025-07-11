package safe

// Note that it must not be embedded, due to the Lock and Unlock methods.
// noCopy prevents a struct from being copied after the first use. It achieves
// this by implementing the [sync.Locker] interface, which triggers the go vet
// copylocks check. It should not be embedded.
//
// https://golang.org/issues/8005#issuecomment-190753527
type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}
