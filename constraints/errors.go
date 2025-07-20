package constraints

type RuntimeError interface {
	error
	RuntimeError()
}
