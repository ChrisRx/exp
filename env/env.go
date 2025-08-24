package env

import (
	"cmp"
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
	"unicode"

	"go.chrisrx.dev/x/expr"
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

// ParseFor parses tags for the provided struct and sets values from environment
// variables. It accepts a struct or pointer to a struct. If a non-pointer, a
// copy of the struct with values set will be returned.
func ParseFor[T any](opts ...ParserOption) (T, error) {
	rv := reflect.New(reflect.TypeFor[T]()).Elem()
	if rv.Kind() != reflect.Pointer {
		v := rv.Interface().(T)
		if err := Parse(&v, opts...); err != nil {
			return v, err
		}
		return v, nil
	}
	v, ok := typeAssert[T](rv)
	if !ok {
		return v, fmt.Errorf("invalid type: %v", v)
	}
	if err := Parse(v, opts...); err != nil {
		return v, err
	}
	return v, nil
}

func typeAssert[T any](rv reflect.Value) (T, bool) {
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

// MustParseFor is a convenience function for calling [ParseFor] that panics if
// an error is encountered.
func MustParseFor[T any](opts ...ParserOption) T {
	v, err := ParseFor[T](opts...)
	if err != nil {
		panic(err)
	}
	return v
}

type ParserOption func(*Parser)

func DisableAutoPrefix() ParserOption {
	return func(p *Parser) {
		p.DisableAutoPrefix = true
	}
}

func Namespace(ns string) ParserOption {
	return func(p *Parser) {
		p.Namespace = ns
	}
}

func RequireTagged() ParserOption {
	return func(p *Parser) {
		p.RequireTagged = true
	}
}

type Parser struct {
	DisableAutoPrefix bool
	Namespace         string
	RequireTagged     bool
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
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}

	// The parser namespace should be added to the initial fields if it is set to
	// ensure the prefix is set for all child fields.
	for i := range rv.NumField() {
		if err := p.parse(rv.Field(i), NewField(rv, i, p.Namespace)); err != nil {
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
		case !p.DisableAutoPrefix && !field.Anonymous:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace, strings.ToSnakeCase(field.Name)))
		default:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace))
		}
		for i := range rv.NumField() {
			if err := p.parse(rv.Field(i), NewField(rv, i, prefixes...)); err != nil {
				return err
			}
		}
		return nil
	default:
		if !p.RequireTagged && field.Env == "" {
			return nil
		}
		if !isValidEnv(field.Env) {
			return fmt.Errorf("env tag must only contain letters, digits or _: %q", field.Env)
		}
		if err := field.Set(rv); err != nil {
			return err
		}
		if field.Validate != "" {
			result, err := expr.Eval(field.Validate, expr.Env(map[string]reflect.Value{
				field.Name: rv,
			}))
			if err != nil {
				return err
			}
			result = underlying(result)
			if result.Kind() != reflect.Bool {
				return fmt.Errorf("expected bool, received %v", result.Type())
			}
			if !result.Bool() {
				return fmt.Errorf(strings.Dedent(`
					field %v failed validation:
						condition: %v
						value: %q
				`), field.Name, field.Validate, rv.Interface())
			}
		}
		return nil
	}
}

func isValidEnv(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
}

// isStruct checks if the provided value is a struct that doesn't have a custom
// parser or implement a common interface. This is important to prevent ranging
// over fields for types like [time.Time], where it is a struct but the fields
// aren't meaningful and we have a parser registered that already knows how to
// handle it optimally.
func isStruct(rv reflect.Value) bool {
	if rv.Type().Kind() != reflect.Struct {
		return false
	}
	if _, ok := customParserFuncs[indirectType(rv.Type())]; ok {
		return false
	}
	if _, ok := typeAssert[encoding.TextUnmarshaler](rv); ok {
		return false
	}
	return true
}

func underlying(rv reflect.Value) reflect.Value {
	if rv.Kind() != reflect.Interface {
		return rv
	}
	if !rv.CanInterface() {
		panic(fmt.Errorf("underlying: cannot interface: %v", rv.Type()))
	}
	return reflect.ValueOf(rv.Interface())
}

type Field struct {
	Name      string
	Anonymous bool

	// tags
	Env         string
	Namespace   string
	Default     string
	DefaultExpr string
	Validate    string
	Separator   string
	Required    bool
	Layout      string

	prefixes []string
}

func NewField(rv reflect.Value, index int, prefixes ...string) Field {
	st := rv.Type().Field(index)
	return Field{
		Name:        st.Name,
		Anonymous:   st.Anonymous,
		Env:         st.Tag.Get("env"),
		Namespace:   st.Tag.Get("namespace"),
		Default:     st.Tag.Get("default"),
		DefaultExpr: st.Tag.Get("$default"),
		Validate:    st.Tag.Get("validate"),
		Separator:   cmp.Or(st.Tag.Get("sep"), ","),
		Required:    must.Get0(strconv.ParseBool(st.Tag.Get("required"))),
		Layout:      cmp.Or(st.Tag.Get("layout"), time.RFC3339Nano),
		prefixes:    slices.Map(slices.DeleteFunc(prefixes, ptr.IsZero), strings.ToUpper),
	}
}

func (f Field) Key() string {
	return strings.Join(append(f.prefixes, f.Env), "_")
}

func (f Field) Set(rv reflect.Value) error {
	s, ok := os.LookupEnv(f.Key())
	if !ok {
		if f.DefaultExpr != "" {
			v, err := expr.Eval(f.DefaultExpr)
			if err != nil {
				return err
			}
			if v.CanConvert(rv.Type()) {
				v = v.Convert(rv.Type())
			}
			if v.Type().PkgPath() == "time" && v.Type().Name() == "Time" && rv.Kind() == reflect.String {
				if method := v.MethodByName("Format"); method.IsValid() {
					if results := method.Call([]reflect.Value{reflect.ValueOf(f.Layout)}); len(results) > 0 {
						v = results[0]
					}
				}
			}
			rv.Set(v)
			return nil
		}
		if f.Default != "" {
			return f.set(rv, f.Default)
		}
		if f.Required {
			return fmt.Errorf("required field not set: %v", f.Key())
		}
		return nil
	}
	return f.set(rv, s)
}

func (f Field) set(rv reflect.Value, s string) error {
	if !rv.CanSet() {
		panic(fmt.Errorf("cannot set value: %v", rv.Type()))
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return f.set(reflect.Indirect(rv), s)
	}

	// When a type-specific parser function is available, this is preferred to
	// continuing default parsing.
	if fn, ok := customParserFuncs[indirectType(rv.Type())]; ok {
		v, err := fn(f, s)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(v))
		return nil
	}

	// Common interfaces, like [encoding.TextUnmarshaler], can be used to set the
	// field value.
	if iface, ok := typeAssert[encoding.TextUnmarshaler](rv); ok {
		return iface.UnmarshalText([]byte(s))
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		et := rv.Type().Elem()
		if isInvalidNestedType(et) {
			return fmt.Errorf("received invalid slice element type: %v", et)
		}
		elems := strings.Split(s, f.Separator)
		sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
		for i := range sv.Len() {
			if err := f.set(sv.Index(i), elems[i]); err != nil {
				return err
			}
		}
		rv.Set(sv)
		return nil
	case reflect.Map:
		elems := strings.Split(s, f.Separator)
		mv := reflect.MakeMap(rv.Type())
		for _, elem := range elems {
			parts := strings.SplitN(elem, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("cannot parse value into key/value pairs: %q", elem)
			}
			key := reflect.New(rv.Type().Key()).Elem()
			if err := f.set(key, parts[0]); err != nil {
				return err
			}
			if isInvalidNestedType(key.Type()) {
				return fmt.Errorf("received invalid map key type: %v", key.Type())
			}
			value := reflect.New(rv.Type().Elem()).Elem()
			if err := f.set(value, parts[1]); err != nil {
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

func isInvalidNestedType(rt reflect.Type) bool {
	switch indirectType(rt).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

func indirectType(rt reflect.Type) reflect.Type {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	return rt
}
