package group

type GroupOption func(*options)

// WithLimit sets the bounded concurrency for a pool of goroutines.
func WithLimit(n int) GroupOption {
	return func(o *options) {
		o.Limit = n
	}
}

const defaultResultsBuffer = 1000

// WithResultsBuffer sets the capacity of the buffered channel used for sending
// results. This option only applies to [ResultGroup].
func WithResultsBuffer(n int) GroupOption {
	return func(o *options) {
		o.ResultsBuffer = n
	}
}

type options struct {
	Limit         int
	ResultsBuffer int
}

func newOptions() *options {
	return &options{
		ResultsBuffer: defaultResultsBuffer,
	}
}

func (o *options) Apply(opts []GroupOption) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}
