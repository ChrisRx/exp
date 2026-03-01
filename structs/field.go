package structs

import (
	"cmp"
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"

	// default built-in conversions
	_ "go.chrisrx.dev/x/structs/conversions"

	"go.chrisrx.dev/x/convert"
	"go.chrisrx.dev/x/expr"
	"go.chrisrx.dev/x/internal/reflectx"
	"go.chrisrx.dev/x/strings"
)

// Field represents a parsed struct field.
type Field reflect.StructField

func (f Field) Default() string {
	return f.Tag.Get("default")
}

func (f Field) DefaultExpr() string {
	return f.Tag.Get("$default")
}

func (f Field) Layout() string {
	return cmp.Or(f.Tag.Get("layout"), time.RFC3339Nano)
}

func (f Field) Separator() string {
	return cmp.Or(f.Tag.Get("sep"), ",")
}

func (f Field) IsExported() bool {
	return reflect.StructField(f).IsExported()
}

func (f Field) HasDefault() bool {
	return f.DefaultExpr() != "" || f.Default() != ""
}

func (f Field) SetDefault(rv reflect.Value) (bool, error) {
	if rv.IsValid() && !rv.IsZero() {
		return false, nil
	}
	opts := []convert.Option{
		convert.Layout(f.Layout()),
		convert.Separator(f.Separator()),
	}
	switch {
	case f.DefaultExpr() != "":
		v, err := expr.Eval(f.DefaultExpr())
		if err != nil {
			return false, err
		}
		if v.CanConvert(rv.Type()) {
			v = v.Convert(rv.Type())
		}
		if fn, ok := convert.Lookup(v.Type(), rv.Type()); ok {
			result, err := fn(v.Interface(), opts...)
			if err != nil {
				return false, err
			}
			rv.Set(reflectx.Indirect(reflect.ValueOf(result), v.Type()))
			return true, nil
		}
		rv.Set(v)
		return true, nil
	case f.Default() != "":
		if err := ParseField(f.Default(), rv, opts...); err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, nil
	}
}

func ParseFieldAs[T any](s string, opts ...convert.Option) (T, error) {
	var zero T
	rv := reflect.New(reflect.TypeFor[T]()).Elem()
	if err := ParseField(s, rv, opts...); err != nil {
		return zero, err
	}
	v, ok := rv.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("invalid type: expected %T, received %v", zero, rv.Type())
	}
	return v, nil
}

func ParseField(s string, rv reflect.Value, opts ...convert.Option) error {
	return (&parser{
		opts: convert.NewOptions(opts),
	}).parse(s, rv)
}

type parser struct {
	opts *convert.Options
}

func (p *parser) parse(s string, rv reflect.Value) error {
	if !rv.CanSet() {
		panic(fmt.Errorf("cannot set value: %v", rv.Type()))
	}

	// When a type-specific parser function is available, this is preferred to
	// continuing default parsing.
	if fn, ok := convert.Lookup(stringType, rv.Type()); ok {
		v, err := fn(s, p.opts.Values()...)
		if err != nil {
			return err
		}
		rv.Set(reflectx.Indirect(reflect.ValueOf(v), rv.Type()))
		return nil
	}

	// Common interfaces, like [encoding.TextUnmarshaler], can be used to set the
	// field value.
	if fn, ok := WellKnownInterfaces(rv); ok {
		return fn(s)
	}

	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parse(s, reflect.Indirect(rv))
	case reflect.Array, reflect.Slice:
		et := rv.Type().Elem()
		if isInvalidNestedType(et) {
			return fmt.Errorf("received invalid slice element type: %v", et)
		}
		elems := strings.Split(s, p.opts.Separator)
		sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
		for i := range sv.Len() {
			if err := p.parse(elems[i], sv.Index(i)); err != nil {
				return err
			}
		}
		rv.Set(sv)
		return nil
	case reflect.Map:
		elems := strings.Split(s, p.opts.Separator)
		mv := reflect.MakeMap(rv.Type())
		for _, elem := range elems {
			parts := strings.SplitN(elem, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("cannot parse value into key/value pairs: %q", elem)
			}
			key := reflect.New(rv.Type().Key()).Elem()
			if err := p.parse(parts[0], key); err != nil {
				return err
			}
			if isInvalidNestedType(key.Type()) {
				return fmt.Errorf("received invalid map key type: %v", key.Type())
			}
			value := reflect.New(rv.Type().Elem()).Elem()
			if err := p.parse(parts[1], value); err != nil {
				return err
			}
			if isInvalidNestedType(value.Type()) {
				return fmt.Errorf("received invalid map value type: %v", value.Type())
			}
			mv.SetMapIndex(key, value)
		}
		rv.Set(mv)
		return nil
	case reflect.String:
		rv.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			return err
		}
		rv.SetUint(i)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		rv.SetBool(b)
		return nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil
	default:
		return fmt.Errorf("received unhandled value: %T", rv.Interface())
	}
}

func isInvalidNestedType(rt reflect.Type) bool {
	switch reflectx.IndirectType(rt).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

var stringType = reflect.TypeFor[string]()

func HasConversion(rv reflect.Value) bool {
	_, ok := convert.Lookup(stringType, rv.Type())
	return ok
}

type WellKnownInterfaceFunc func(s string) error

func WellKnownInterfaces(rv reflect.Value) (WellKnownInterfaceFunc, bool) {
	if v := reflectx.Interface(rv); v != nil {
		switch v := v.(type) {
		case encoding.TextUnmarshaler:
			return func(s string) error {
				return v.UnmarshalText([]byte(s))
			}, true
		}
	}
	return nil, false
}

func IsWellKnown(rv reflect.Value) bool {
	_, ok := WellKnownInterfaces(rv)
	return ok
}
