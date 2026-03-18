package archive

type options struct {
	Concurrency   int
	Limit, Offset int
}

type Option func(*options)

func WithConcurrency(n int) Option {
	return func(o *options) {
		o.Concurrency = n
	}
}

func WithLimit(n int) Option {
	return func(o *options) {
		o.Limit = n
	}
}

func WithOffset(n int) Option {
	return func(o *options) {
		o.Offset = n
	}
}
