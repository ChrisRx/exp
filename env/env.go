package env

import (
	"cmp"
	"encoding"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"time"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/strings"
)

// Parse parses tags for the provided struct and sets values from environment
// variables. It will only set values that have the `env` tag. Value must be a
// pointer value to allow setting values.
func Parse(v any, opts ...ParserOption) error {
	return NewParser(opts...).Parse(v)
}

// Parse is a convenience function for calling [Parse] that panics if an error
// is encountered.
func MustParse(v any, opts ...ParserOption) {
	if err := Parse(v, opts...); err != nil {
		panic(err)
	}
}

// ParseAs parses tags for the provided struct and sets values from environment
// variables. It accepts a struct or pointer to a struct. If a non-pointer, a
// copy of the struct with values set will be returned.
func ParseAs[T any](opts ...ParserOption) (T, error) {
	rv := reflect.New(reflect.TypeFor[T]()).Elem()
	if rv.Kind() != reflect.Pointer {
		v := rv.Interface().(T)
		if err := Parse(&v, opts...); err != nil {
			return v, err
		}
		return v, nil
	}
	v, ok := TypeAssert[T](rv)
	if !ok {
		var zero T
		return zero, fmt.Errorf("invalid type: %v", v)
	}
	if err := Parse(v, opts...); err != nil {
		return v, err
	}
	return v, nil
}

// MustParseAs is a convenience function for calling [ParseAs] that panics if
// an error is encountered.
func MustParseAs[T any](opts ...ParserOption) T {
	v, err := ParseAs[T](opts...)
	if err != nil {
		panic(err)
	}
	return v
}

type ParserOption func(*Parser)

func RequireTagged() ParserOption {
	return func(p *Parser) {
		p.RequireTagged = true
	}
}

func DisableAutoTag() ParserOption {
	return func(p *Parser) {
		p.DisableAutoTag = true
	}
}

func Separator(sep string) ParserOption {
	return func(p *Parser) {
		p.Separator = sep
	}
}

type Parser struct {
	Namespace      string
	RequireTagged  bool
	DisableAutoTag bool
	Separator      string
}

func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{
		Separator: ",",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Parser) Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("must provide a struct or *struct, received %T", v)
	}
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct or *struct, received %T", v)
	}

	// The parser namespace should be added to the initial fields if it is set to
	// ensure the prefix is set for all child fields.
	for i := range rv.NumField() {
		field := GetField(rv.Type().Field(i), p.Namespace)
		if !p.RequireTagged && field.ShouldSkip() {
			return nil
		}
		if err := p.parse(rv.Field(i), field); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parse(rv reflect.Value, field Field) error {
	switch {
	case rv.Kind() == reflect.Pointer:
		if rv.IsNil() {
			// A nil pointer must have the underlying type initialized to be
			// settable.
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parse(reflect.Indirect(rv), field)
	case isStruct(rv):
		// Any prefixes from the parent field should be added to child fields. An
		// additional prefix will added if either env or namespace tags are
		// set, or if auto prefix is not disabled and the parent field wasn't
		// anonymous (aka embedded).
		prefixes := field.prefixes
		switch {
		case !p.DisableAutoTag && !field.Anonymous:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace, strings.ToSnakeCase(field.Name)))
		default:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace))
		}
		for i := range rv.NumField() {
			field := GetField(rv.Type().Field(i), prefixes...)
			if !p.RequireTagged && field.ShouldSkip() {
				return nil
			}
			if err := p.parse(rv.Field(i), field); err != nil {
				return err
			}
		}
		return nil
	default:
		s, ok := field.Get()
		if !ok || s == "" {
			if p.RequireTagged || field.Required {
				return fmt.Errorf("required field not set: %v", field.Key())
			}
			return nil
		}
		return p.set(rv, field, s)
	}
}

func isStruct(rv reflect.Value) bool {
	if rv.Type().Kind() != reflect.Struct {
		return false
	}
	if _, ok := LookupFunc(rv.Type()); ok {
		return false
	}
	if _, ok := TypeAssert[encoding.TextUnmarshaler](rv); ok {
		return false
	}
	return true
}

func (p *Parser) set(rv reflect.Value, field Field, s string) error {
	if !rv.CanSet() {
		panic(fmt.Errorf("cannot set value: %v", rv.Type()))
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.set(reflect.Indirect(rv), field, s)
	}

	// When a type-specific parser function is available, this is preferred to
	// continuing default parsing.
	if fn, ok := LookupFunc(rv.Type()); ok {
		v, err := fn(s, field)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(v))
		return nil
	}

	// Common interfaces, like [encoding.TextUnmarshaler], can be used to set the
	// field value.
	if iface, ok := TypeAssert[encoding.TextUnmarshaler](rv); ok {
		return iface.UnmarshalText([]byte(s))
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		et := rv.Type().Elem()
		if !isSlice(et) {
			// Slices of slices are not supported.
			return fmt.Errorf("received invalid slice element type: %v", et)
		}
		elems := strings.Split(s, p.Separator)
		sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
		for i := range sv.Len() {
			if err := p.set(sv.Index(i), field, elems[i]); err != nil {
				return err
			}
		}
		rv.Set(sv)
		return nil
	case reflect.Map:
		elems := strings.Split(s, p.Separator)
		mv := reflect.MakeMap(rv.Type())
		for _, elem := range elems {
			parts := strings.SplitN(elem, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("cannot parse value into key/value pairs: %q", elem)
			}
			key := reflect.New(rv.Type().Key()).Elem()
			if err := p.set(key, field, parts[0]); err != nil {
				return err
			}
			value := reflect.New(rv.Type().Elem()).Elem()
			if err := p.set(value, field, parts[1]); err != nil {
				return err
			}
			mv.SetMapIndex(key, value)
		}
		rv.Set(mv)
		return nil
	case reflect.String:
		rv.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 64)
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

func isSlice(rt reflect.Type) bool {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice {
		return false
	}
	return true
}

type Field struct {
	Name      string
	Type      reflect.Type
	Anonymous bool

	Env       string
	Namespace string
	Default   string
	Required  bool
	Ignored   bool
	Layout    string

	prefixes []string
}

func GetField(st reflect.StructField, prefixes ...string) Field {
	return Field{
		Name:      st.Name,
		Type:      IndirectType(st.Type),
		Anonymous: st.Anonymous,
		Env:       st.Tag.Get("env"),
		Namespace: st.Tag.Get("namespace"),
		Default:   st.Tag.Get("default"),
		Required:  must.Get0(strconv.ParseBool(st.Tag.Get("required"))),
		Ignored:   st.Tag.Get("env") == "-",
		Layout:    cmp.Or(st.Tag.Get("layout"), time.RFC3339Nano),
		prefixes:  slices.Map(slices.DeleteFunc(prefixes, ptr.IsZero), strings.ToUpper),
	}
}

func (f Field) Get() (s string, ok bool) {
	slog.Debug("get env value", slog.String("key", f.Key()))
	if s, ok := os.LookupEnv(f.Key()); ok {
		return s, true
	}
	return f.Default, f.Default != ""
}

func (f Field) Key() string {
	return strings.Join(append(f.prefixes, f.Env), "_")
}

func (f Field) ShouldSkip() bool {
	return f.Type.Kind() != reflect.Struct && (f.Ignored || f.Env == "")
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

func IndirectType(rt reflect.Type) reflect.Type {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	return rt
}
