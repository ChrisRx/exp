// Package env provides functions for loading structs from environment
// variables.
package env

import (
	"cmp"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unicode"

	"go.chrisrx.dev/x/convert"
	"go.chrisrx.dev/x/expr"
	"go.chrisrx.dev/x/internal/reflectx"
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/strings"
	"go.chrisrx.dev/x/structs"
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
	v, ok := reflectx.TypeAssert[T](rv)
	if !ok {
		return v, fmt.Errorf("invalid type: %v", v)
	}
	if err := Parse(v, opts...); err != nil {
		return v, err
	}
	return v, nil
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

// Deferred is a special type used in fields to configure calling methods on
// structs after parsing is done.
type Deferred bool

func isDeferred(rv reflect.Value) bool {
	return rv.Type() == reflect.TypeFor[Deferred]()
}

type ParserOption func(*Parser)

// DisableAutoPrefix is an option for [Parser] that disables the auto-prefix
// feature.
func DisableAutoPrefix() ParserOption {
	return func(p *Parser) {
		p.DisableAutoPrefix = true
	}
}

// RequireTagged is an option for [Parser] that makes all struct fields
// required. This applies regardless of whether the `env` tag is set.
func RequireTagged() ParserOption {
	return func(p *Parser) {
		p.RequireTagged = true
	}
}

// RootPrefix is an option for [Parser] that sets a root prefix for environment
// variable names.
func RootPrefix(prefix string) ParserOption {
	return func(p *Parser) {
		p.RootPrefix = prefix
	}
}

// Parser is an environment variable parser for structs.
type Parser struct {
	DisableAutoPrefix bool
	RootPrefix        string
	RequireTagged     bool

	deferred []string
}

// NewParser constructs a new [Parser] using the provided options.
func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Parse loads values into a struct from environment variables. It only accepts
// a pointer to a struct.
func (p *Parser) Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}
	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}

	// The parser root prefix should be added to the initial fields if it is set to
	// ensure the prefix is set for all child fields.
	for i := range rv.NumField() {
		if err := p.parse(rv.Field(i), newField(rv.Type().Field(i), p.RootPrefix)); err != nil {
			return err
		}
	}
	for _, deferred := range p.deferred {
		method := rv.MethodByName(deferred)
		if !method.IsValid() {
			return fmt.Errorf("deferred method is invalid: %q", deferred)
		}
		if method.Type().NumIn() > 0 {
			return fmt.Errorf("deferred method cannot accept arguments")
		}
		method.Call(nil)
	}
	return nil
}

func (p *Parser) parse(rv reflect.Value, field Field) error {
	if !field.IsExported() && !isDeferred(rv) {
		return nil
	}

	switch {
	case rv.Kind() == reflect.Pointer:
		if rv.IsNil() {
			// A nil pointer must have the underlying type initialized to be
			// settable.
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parse(reflect.Indirect(rv), field)
	case isDeferred(rv):
		s, ok := os.LookupEnv(field.Key())
		if !ok {
			if field.Default() == "" {
				return nil
			}
			s = field.Default()
		}
		if must.Get0(strconv.ParseBool(s)) {
			p.deferred = append(p.deferred, field.DeferredMethod)
		}
		return nil
	case structs.HasConversion(rv), structs.IsWellKnown(rv):
		// If we have a custom parser or a common interface, it is important that
		// we don't range over it as a struct, so we go ahead and parse it as a
		// singular value.
		return p.parseSingular(rv, field)
	case rv.Kind() == reflect.Struct:
		// Any prefixes from the parent field should be added to child fields. An
		// additional prefix will added if the env tag is set, or if auto prefix is
		// not disabled and the parent field wasn't anonymous (aka embedded).
		prefixes := field.prefixes
		switch {
		case !p.DisableAutoPrefix && !field.Anonymous:
			prefixes = append(prefixes, cmp.Or(field.Env, strings.ToSnakeCase(field.Name)))
		default:
			prefixes = append(prefixes, field.Env)
		}
		for i := range rv.NumField() {
			if err := p.parse(rv.Field(i), newField(rv.Type().Field(i), prefixes...)); err != nil {
				return err
			}
		}
		return nil
	default:
		return p.parseSingular(rv, field)
	}
}

func (p *Parser) parseSingular(rv reflect.Value, field Field) error {
	if !p.RequireTagged && field.Env == "" {
		return nil
	}
	if !isValidEnv(field.Env) {
		return fmt.Errorf("env tag must only contain letters, digits or _: %q", field.Env)
	}
	if err := field.set(rv); err != nil {
		return err
	}
	if field.Validate != "" {
		result, err := expr.Eval(field.Validate, expr.Env(map[string]reflect.Value{
			field.Name: rv,
			"self":     rv,
		}))
		if err != nil {
			return err
		}
		result = reflectx.Underlying(result)
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

func isValidEnv(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
}

type CustomParserFunc[T string, R any] = convert.ConversionFunc[T, R]

// Register registers a custom parser with the provided type parameter. The
// type parameter must be a non-pointer type, however, registering a type will
// match for parsing for both the pointer and non-pointer of the type.
func Register[T any](fn CustomParserFunc[string, T]) {
	convert.Register(fn)
}
