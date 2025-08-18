package env

import (
	"cmp"
	"encoding"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strconv"
	"time"

	"go.chrisrx.dev/x/must"
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
func MustParse[T any](v T, opts ...ParserOption) {
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
	rv = reflect.New(reflect.TypeFor[T]().Elem())
	v, ok := rv.Interface().(T)
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

type Parser struct {
	Namespace      string
	RequireTagged  bool
	DisableAutoTag bool
}

func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Parser) Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || reflect.Indirect(rv).Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to a struct, received %T", v)
	}
	rv = reflect.Indirect(rv)

	var ns Namespace
	if p.Namespace != "" {
		ns = append(ns, p.Namespace)
	}

	for i := range rv.NumField() {
		fv := rv.Field(i)
		field := GetField(rv.Type().Field(i))
		if err := p.parse(fv, ns, field); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parse(rv reflect.Value, ns Namespace, field Field) error {
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parse(reflect.Indirect(rv), ns, field)
	}

	// First, check for customer parsers. This allows us to short circuit if we
	// have a specific parser function. It is important that this checks first to
	// enable overriding any default parsing for a value.
	if fn, ok := Lookup(rv.Type()); ok {
		s, err := p.get(ns, field)
		if err != nil {
			return err
		}
		v, err := fn(s, field)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(v))
		return nil
	}
	// Check for common interfaces that would allow us to avoid additional
	// parsing. Anything that implements [encoding.TextUnmarshaler] can be used
	// to set the value.
	if v, ok := TypeAssert[encoding.TextUnmarshaler](rv); ok {
		s, err := p.get(ns, field)
		if err != nil {
			return err
		}
		if err := v.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		return nil
	}

	switch rv.Kind() {
	case reflect.Struct:
		if !p.DisableAutoTag {
			ns = ns.Append(field)
		}
		rv = reflect.Indirect(rv)
		for i := range rv.NumField() {
			fv := rv.Field(i)
			field := GetField(rv.Type().Field(i))
			if err := p.parse(fv, ns, field); err != nil {
				return err
			}
		}
		return nil
	default:
		if !p.DisableAutoTag {
			ns = ns.Append(field)
		}
		v, err := p.get(ns, field)
		if err != nil {
			return err
		}
		if v == "" {
			return nil
		}
		return p.set(rv, v)
	}
}

func (p *Parser) get(ns Namespace, field Field) (string, error) {
	if field.Ignored || field.Env == "" {
		return "", nil
	}
	slog.Debug("get env value", slog.String("key", ns.Join(field.Env)))
	if v, ok := os.LookupEnv(ns.Join(field.Env)); ok {
		return v, nil
	}
	if field.Default != "" {
		return field.Default, nil
	}
	if p.RequireTagged || field.Required {
		return "", fmt.Errorf("required field not set: %v", ns.Join(field.Env))
	}
	return "", nil
}

func (p *Parser) set(rv reflect.Value, v string) error {
	if !rv.CanSet() {
		panic(fmt.Errorf("cannot set value: %v", rv.Type()))
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(i)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		rv.SetBool(b)
		return nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil
	case reflect.Array, reflect.Slice:
		return p.setSlice(rv, v)
	case reflect.Pointer:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.set(reflect.Indirect(rv), v)
	default:
		return fmt.Errorf("setValue: received %T", rv.Interface())
	}
}

var allowedSliceElementKinds = []reflect.Kind{
	reflect.String,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
	reflect.Pointer,
}

func (p *Parser) setSlice(rv reflect.Value, v string) error {
	et := rv.Type().Elem()
	if !slices.Contains(allowedSliceElementKinds, et.Kind()) {
		return fmt.Errorf("received invalid slice element type: %v", et)
	}
	elems := slices.DeleteFunc(strings.Split(v, ","), func(s string) bool {
		return s == ""
	})
	if len(elems) == 0 {
		return nil
	}
	sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
	for i := range sv.Len() {
		if err := p.set(sv.Index(i), elems[i]); err != nil {
			return err
		}
	}
	rv.Set(sv)
	return nil
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
}

func GetField(ft reflect.StructField) Field {
	name := ft.Name
	rt := ft.Type
	if ft.Type.Kind() == reflect.Pointer {
		rt = ft.Type.Elem()
	}
	return Field{
		Name:      name,
		Type:      rt,
		Anonymous: ft.Anonymous,
		Env:       ft.Tag.Get("env"),
		Namespace: ft.Tag.Get("namespace"),
		Default:   ft.Tag.Get("default"),
		Required:  must.Get0(strconv.ParseBool(ft.Tag.Get("required"))),
		Ignored:   ft.Tag.Get("env") == "-",
		Layout:    cmp.Or(ft.Tag.Get("layout"), time.RFC3339Nano),
	}
}

type Namespace []string

func (n Namespace) Append(field Field) Namespace {
	if field.Type.Kind() == reflect.Struct && !field.Anonymous {
		ns := cmp.Or(field.Env, field.Namespace)
		if ns == "" {
			ns = strings.ToSnakeCase(field.Name)
		}
		n = append(n, strings.ToUpper(ns))
	}
	return n
}

func (n Namespace) Join(elems ...string) string {
	return strings.Join(append(n, elems...), "_")
}

func (n Namespace) String() string {
	return strings.Join(n, "_")
}

func (n *Namespace) Pop() {
	if len(*n) == 0 {
		return
	}
	*n = (*n)[:len(*n)-1]
}

func TypeAssert[T any](rv reflect.Value) (T, bool) {
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

func MakeAddressable(v any) reflect.Value {
	rv := reflect.New(reflect.TypeOf(v))
	rv.Elem().Set(reflect.ValueOf(v))
	return rv
}
