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
	if rv.Kind() != reflect.Pointer || reflect.Indirect(rv).Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to a struct, received %T", v)
	}
	rv = reflect.Indirect(rv)
	for i := range rv.NumField() {
		fv := rv.Field(i)
		field := GetField(rv.Type().Field(i))
		if !p.RequireTagged && field.ShouldSkip() {
			return nil
		}
		if p.Namespace != "" {
			field.prefixes = append(field.prefixes, p.Namespace)
		}
		if err := p.parse(fv, field); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parse(rv reflect.Value, field Field) error {
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parse(reflect.Indirect(rv), field)
	}

	if isStruct(rv) {
		// Add to the prefix when a field is not embedded. If auto tagging is
		// disabled then no prefixes are added. If the env or namespace tag is set,
		// those are preferred, otherwise it uses the parent field name.
		prefixes := field.prefixes
		if !p.DisableAutoTag && !field.Anonymous {
			ns := cmp.Or(field.Env, field.Namespace, strings.ToSnakeCase(field.Name))
			if ns != "" {
				prefixes = append(prefixes, strings.ToUpper(ns))
			}
		}
		for i := range reflect.Indirect(rv).NumField() {
			fv := rv.Field(i)
			field := GetField(rv.Type().Field(i))
			if !p.RequireTagged && field.ShouldSkip() {
				return nil
			}
			field.prefixes = append(field.prefixes, prefixes...)
			if err := p.parse(fv, field); err != nil {
				return err
			}
		}
		return nil
	}

	s, ok := field.Get()
	if !ok || s == "" {
		if p.RequireTagged || field.Required {
			return fmt.Errorf("required field not set: %v", field.Key())
		}
		return nil
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		et := rv.Type().Elem()
		if !isSlice(et) {
			return fmt.Errorf("received invalid slice element type: %v", et)
		}
		elems := slices.DeleteFunc(strings.Split(s, p.Separator), func(s string) bool {
			return s == ""
		})
		if len(elems) == 0 {
			return nil
		}
		sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
		for i := range sv.Len() {
			if err := p.set(sv.Index(i), elems[i], field); err != nil {
				return err
			}
		}
		rv.Set(sv)
		return nil
	default:
		return p.set(rv, s, field)
	}
}

func isStruct(rv reflect.Value) bool {
	if rv.Type().Kind() != reflect.Struct {
		return false
	}
	if _, ok := Lookup(rv.Type()); ok {
		return false
	}
	if _, ok := TypeAssert[encoding.TextUnmarshaler](rv); ok {
		return false
	}
	return true
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

func (p *Parser) set(rv reflect.Value, s string, field Field) error {
	if !rv.CanSet() {
		panic(fmt.Errorf("cannot set value: %v", rv.Type()))
	}

	// First, check for customer parsers. This allows us to short circuit if we
	// have a specific parser function. It is important that this checks first to
	// enable overriding any default parsing for a value.
	if fn, ok := Lookup(rv.Type()); ok {
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
	if iface, ok := TypeAssert[encoding.TextUnmarshaler](rv); ok {
		if err := iface.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		return nil
	}

	switch rv.Kind() {
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
	case reflect.Pointer:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.set(reflect.Indirect(rv), s, field)
	default:
		return fmt.Errorf("received unhandled value: %T", rv.Interface())
	}
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

func GetField(ft reflect.StructField) Field {
	return Field{
		Name:      ft.Name,
		Type:      IndirectType(ft.Type),
		Anonymous: ft.Anonymous,
		Env:       ft.Tag.Get("env"),
		Namespace: ft.Tag.Get("namespace"),
		Default:   ft.Tag.Get("default"),
		Required:  must.Get0(strconv.ParseBool(ft.Tag.Get("required"))),
		Ignored:   ft.Tag.Get("env") == "-",
		Layout:    cmp.Or(ft.Tag.Get("layout"), time.RFC3339Nano),
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
