package convert

import (
	"reflect"

	"go.chrisrx.dev/x/internal/reflectx"
)

type ConversionFunc[From, To any] func(From, ...Option) (To, error)

type Registry map[[2]reflect.Type]ConversionFunc[any, any]

// Register registers a custom parser with the provided type parameter. The
// type parameter must be a non-pointer type, however, registering a type will
// match for parsing for both the pointer and non-pointer of the type.
func (r Registry) Register[From, To any](fn ConversionFunc[From, To]) {
	r[convertFor[From, To]()] = func(v any, opts ...Option) (any, error) {
		in, err := reflectx.IndirectFor[From](v)
		if err != nil {
			return nil, err
		}
		result, err := fn(in, opts...)
		if err != nil {
			return nil, err
		}
		return reflectx.IndirectFor[To](result)
	}
}

func (r Registry) Lookup(from, to reflect.Type) (ConversionFunc[any, any], bool) {
	fn, ok := r[convert(from, to)]
	return fn, ok
}

func convert(from, to reflect.Type) [2]reflect.Type {
	return [2]reflect.Type{
		reflectx.IndirectType(from),
		reflectx.IndirectType(to),
	}
}

func (r Registry) LookupFor[From, To any]() (ConversionFunc[From, To], bool) {
	fn, ok := r[convertFor[From, To]()]
	if !ok {
		return nil, false
	}
	return func(v From, opts ...Option) (To, error) {
		result, err := fn(v, opts...)
		if err != nil {
			return *new(To), err
		}
		return reflectx.IndirectFor[To](result)
	}, true
}

func convertFor[From, To any]() [2]reflect.Type {
	return convert(reflect.TypeFor[From](), reflect.TypeFor[To]())
}

// The default registry used by package functions.
var registry = make(Registry)

// Register registers a custom parser with the provided type parameter. The
// type parameter must be a non-pointer type, however, registering a type will
// match for parsing for both the pointer and non-pointer of the type.
func Register[From, To any](fn ConversionFunc[From, To]) {
	registry.Register(fn)
}

func Lookup(from, to reflect.Type) (ConversionFunc[any, any], bool) {
	return registry.Lookup(from, to)
}

func LookupFor[From, To any]() (ConversionFunc[From, To], bool) {
	return registry.LookupFor[From, To]()
}
