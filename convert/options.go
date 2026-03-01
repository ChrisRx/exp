package convert

import "time"

type Options struct {
	Separator string
	Layout    string

	opts []Option
}

func (o Options) Values() []Option {
	return o.opts
}

func NewOptions(opts []Option) *Options {
	o := &Options{
		Separator: ",",
		Layout:    time.RFC3339Nano,
		opts:      opts,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Option func(*Options)

func Layout(layout string) Option {
	return func(o *Options) {
		o.Layout = layout
	}
}

func Separator(sep string) Option {
	return func(o *Options) {
		o.Separator = sep
	}
}
