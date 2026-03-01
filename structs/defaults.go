package structs

import (
	"fmt"
	"reflect"

	"go.chrisrx.dev/x/must"
)

// Defaults applys any defaults defined in struct tags to fields. Default
// values are only applied to fields that have a `default` struct tag. It
// behaves differently depending upon the type and value it is given:
//
// When given a struct pointer, defaults will be applied for all non-zero
// fields of the struct, modifying the pointed to struct in-place. This
// includes any nested struct fields, as well. The struct pointer is also
// returned but will be the same struct pointer passed to Defaults.
//
// When given a struct, defaults are applied in the same way as with the struct
// pointer, but to a newly constructed struct, which is also returned.
func Defaults[T any](v T) T {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsValid() && rv.IsNil() {
			panic("value provided to Defaults must be initialized")
		}
		if err := TryDefaults(v); err != nil {
			panic(err)
		}
		return v
	}
	if err := TryDefaults(&v); err != nil {
		panic(err)
	}
	return v
}

// DefaultsFor applies defaults defined in struct tags to the struct or struct
// pointer specified in the type paramater. It works the same as [Defaults],
// but only requires a type and constructs a new value in all cases.
func DefaultsFor[T any]() T {
	var v T
	if err := TryDefaults(&v); err != nil {
		panic(err)
	}
	return v
}

func TryDefaults(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}

	for i := range rv.NumField() {
		if err := setDefault(rv.Field(i), Field(rv.Type().Field(i))); err != nil {
			return err
		}
	}

	return nil
}

func setDefault(rv reflect.Value, field Field) error {
	switch {
	case rv.Kind() == reflect.Pointer:
		if rv.IsNil() {
			// A nil pointer must have the underlying type initialized to be
			// settable.
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return setDefault(reflect.Indirect(rv), field)
	case HasConversion(rv), IsWellKnown(rv):
		return must.Get1(field.SetDefault(rv))
	case rv.Kind() == reflect.Struct:
		for i := range rv.NumField() {
			if err := setDefault(rv.Field(i), Field(rv.Type().Field(i))); err != nil {
				return err
			}
		}
		return nil
	default:
		return must.Get1(field.SetDefault(rv))
	}
}
