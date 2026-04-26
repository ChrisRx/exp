package reflectx

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func IndirectType(rt reflect.Type) reflect.Type {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	return rt
}

func Interface(rv reflect.Value) any {
	rt := rv.Type()
	if rt.Kind() != reflect.Interface {
		if rv.Type().Kind() == reflect.Pointer && rv.IsNil() {
			rv = reflect.New(rt.Elem())
		}
	}
	if rv.CanAddr() {
		rv = rv.Addr()
	}
	if !rv.CanInterface() {
		rv = reflect.New(rv.Type()).Elem()
	}
	return rv.Interface()
}

func TypeAssert[T any](rv reflect.Value) (T, bool) {
	rt := reflect.TypeFor[T]()
	if rt.Kind() != reflect.Interface {
		if rv.Type().Kind() == reflect.Pointer && rv.IsNil() {
			rv = reflect.New(rt.Elem())
		}
	}
	if rv.CanAddr() {
		rv = rv.Addr()
	}
	if !rv.CanInterface() {
		var zero T
		return zero, false
	}
	v, ok := rv.Interface().(T)
	return v, ok
}

func Underlying(rv reflect.Value) reflect.Value {
	if rv.Kind() != reflect.Interface {
		return rv
	}
	if !rv.CanInterface() {
		panic(fmt.Errorf("underlying: cannot interface: %v", rv.Type()))
	}
	return reflect.ValueOf(rv.Interface())
}

func MakeAddressable(v reflect.Value) reflect.Value {
	if v.CanAddr() {
		return v.Addr()
	}
	rv := reflect.New(v.Type())
	rv.Elem().Set(v)
	return rv
}

func MakeAddressableField(rv reflect.Value, fieldNum int) reflect.Value {
	rv2 := reflect.New(rv.Type()).Elem()
	rv2.Set(rv)
	fv := rv2.Field(fieldNum)
	return reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
}

var ErrInvalidType = errors.New("invalid type for conversion func")

func IndirectFor[T any](v any) (T, error) {
	rv := reflect.ValueOf(v)
	to := reflect.TypeFor[T]()
	switch {
	case to.Kind() == reflect.Pointer && rv.Kind() != reflect.Pointer:
		rv = MakeAddressable(rv)
	case to.Kind() != reflect.Pointer && rv.Kind() == reflect.Pointer:
		rv = rv.Elem()
	}
	out, ok := reflect.TypeAssert[T](rv)
	if !ok {
		return *new(T), fmt.Errorf("%w: expected %T, received %T", ErrInvalidType, out, v)
	}
	return out, nil
}

func Indirect(rv reflect.Value, to reflect.Type) reflect.Value {
	switch {
	case to.Kind() == reflect.Pointer && rv.Kind() != reflect.Pointer:
		rv = MakeAddressable(rv)
	case to.Kind() != reflect.Pointer && rv.Kind() == reflect.Pointer:
		rv = rv.Elem()
	}
	return rv
}
