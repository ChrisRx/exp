package convert

import (
	"reflect"

	"go.chrisrx.dev/x/internal/reflectx"
)

type key struct {
	From, To reflect.Type
}

type conversionFunc func(any, ...Option) (any, error)

// TODO(ChrisRx): When generic methods land it will be much easier to create
// instances of the conversion registry:
//
// https://github.com/golang/go/issues/77273
var conversions = make(map[key]func(any, ...Option) (any, error))

func convert(from, to reflect.Type) key {
	return key{
		From: reflectx.IndirectType(from),
		To:   reflectx.IndirectType(to),
	}
}

func convertFor[I, O any]() key {
	return convert(reflect.TypeFor[I](), reflect.TypeFor[O]())
}

type ConversionFunc[T, R any] func(T, ...Option) (R, error)

// Register registers a custom parser with the provided type parameter. The
// type parameter must be a non-pointer type, however, registering a type will
// match for parsing for both the pointer and non-pointer of the type.
func Register[T, R any](fn ConversionFunc[T, R]) {
	from, to := reflect.TypeFor[T](), reflect.TypeFor[R]()

	conversions[convert(from, to)] = func(v any, opts ...Option) (any, error) {
		in, err := reflectx.IndirectFor[T](v)
		if err != nil {
			return nil, err
		}
		result, err := fn(in, opts...)
		if err != nil {
			return nil, err
		}
		return reflectx.IndirectFor[R](result)
	}
}

func Lookup(from, to reflect.Type) (conversionFunc, bool) {
	fn, ok := conversions[convert(from, to)]
	return fn, ok
}

func LookupFor[T, R any]() (ConversionFunc[T, R], bool) {
	fn, ok := conversions[convertFor[T, R]()]
	if !ok {
		return nil, false
	}
	return func(v T, opts ...Option) (R, error) {
		result, err := fn(v, opts...)
		if err != nil {
			return *new(R), err
		}
		return reflectx.IndirectFor[R](result)
	}, true
}
