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
func ParseAs[T any]() (T, error) {
	v := *new(T)
	if !isPointer(v) {
		if err := Parse(&v); err != nil {
			return v, err
		}
		return v, nil
	}
	if err := Parse(v); err != nil {
		return v, err
	}
	return v, nil
}

// MustParseAs is a convenience function for calling [ParseAs] that panics if
// an error is encountered.
func MustParseAs[T any]() T {
	v, err := ParseAs[T]()
	if err != nil {
		panic(err)
	}
	return v
}

type ParserOption func(*Parser)

type Parser struct {
	Namespace      string
	Required       bool
	DisableAutoTag bool
}

func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func RequireAll() ParserOption {
	return func(p *Parser) {
		p.Required = true
	}
}

func DisableAutoTag() ParserOption {
	return func(p *Parser) {
		p.DisableAutoTag = true
	}
}

func (p *Parser) Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || reflect.Indirect(rv).Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to a struct, received %T", v)
	}
	return p.parseFields(rv, Namespace{})
}

func (p *Parser) parseFields(rv reflect.Value, ns Namespace) error {
	rv = reflect.Indirect(rv)

	for i := range rv.NumField() {
		fv := rv.Field(i)
		ft := rv.Type().Field(i)
		tags := GetFieldTags(ft)

		if fn, ok := customParserFuncs[fv.Type()]; ok {
			s, err := p.get(ns, tags)
			if err != nil {
				return err
			}
			v, err := fn(s, tags)
			if err != nil {
				return err
			}
			fv.Set(reflect.ValueOf(v))
			continue
		}

		if iface := as[encoding.TextUnmarshaler](fv); iface != nil {
			s, err := p.get(ns, tags)
			if err != nil {
				return err
			}
			if err := iface.UnmarshalText([]byte(s)); err != nil {
				return err
			}
			continue
		}

		if err := p.parse(fv, ns.Append(ft), tags); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) parse(rv reflect.Value, ns Namespace, tags FieldTags) error {
	switch rv.Kind() {
	case reflect.Pointer:
		return p.parse(reflect.Indirect(rv), ns, tags)
	case reflect.Struct:
		return p.parseFields(rv, ns)
	default:
		v, err := p.get(ns, tags)
		if err != nil {
			return err
		}
		return p.set(rv, v)
	}
}

func (p *Parser) get(ns Namespace, tags FieldTags) (string, error) {
	if tags.Ignored || tags.Env == "" {
		return "", nil
	}
	slog.Debug("get env value",
		slog.String("key", strings.Join(append(ns, tags.Env), "_")),
	)
	if v, ok := ns.LookupEnv(tags.Env); ok {
		return v, nil
	}
	if tags.Default != "" {
		return tags.Default, nil
	}
	if p.Required || tags.Required {
		return "", fmt.Errorf("required field not set: %v", strings.Join(append(ns, tags.Env), "_"))
	}
	return "", nil
}

func (p *Parser) set(rv reflect.Value, v string) error {
	if !rv.CanSet() {
		return fmt.Errorf("cannot set value: %v", rv)
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
	default:
		return fmt.Errorf("setValue: received %T", rv.Interface())
	}
}

func (p *Parser) setSlice(rv reflect.Value, v string) error {
	et := rv.Type().Elem()
	switch et.Kind() {
	case reflect.String:
		elems := strings.Split(v, ",")
		sv := reflect.MakeSlice(reflect.SliceOf(et), len(elems), len(elems))
		for i := range sv.Len() {
			ev := sv.Index(i)
			ev.SetString(elems[i])
		}
		rv.Set(sv)
		return nil
	default:
		return fmt.Errorf("setSlice: received %v", et)
	}
}

type FieldTags struct {
	Name     string
	Env      string
	Default  string
	Required bool
	Ignored  bool

	Layout string
}

func GetFieldTags(ft reflect.StructField) FieldTags {
	name := ft.Name
	if ft.Type.PkgPath() != "" {
		name = fmt.Sprintf("%s.%s", ft.Type.PkgPath(), ft.Name)
	}
	return FieldTags{
		Name:     name,
		Env:      ft.Tag.Get("env"),
		Default:  ft.Tag.Get("default"),
		Required: must.Get0(strconv.ParseBool(ft.Tag.Get("required"))),
		Ignored:  must.Get0(strconv.ParseBool(ft.Tag.Get("ignored"))),
		Layout:   cmp.Or(ft.Tag.Get("layout"), time.RFC3339Nano),
	}
}

type Namespace []string

func (n Namespace) Append(ft reflect.StructField) Namespace {
	rt := ft.Type
	if ft.Type.Kind() == reflect.Pointer {
		rt = ft.Type.Elem()
	}
	if rt.Kind() == reflect.Struct && !ft.Anonymous {
		ns := ft.Tag.Get("namespace")
		if ns == "" {
			ns = strings.ToSnakeCase(ft.Name)
		}
		n = append(n, strings.ToUpper(ns))
	}
	return n
}

func (n Namespace) LookupEnv(env string) (string, bool) {
	return os.LookupEnv(append(n, env).String())
}

func (n Namespace) String() string {
	return strings.Join(n, "_")
}

func (n *Namespace) Pop() {
	*n = (*n)[:len(*n)-1]
}

func as[T any](rv reflect.Value) T {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
	} else if rv.CanAddr() {
		rv = rv.Addr()
	}

	tm, ok := rv.Interface().(T)
	if !ok {
		var zero T
		return zero
	}
	return tm
}

func isPointer(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Pointer
}
