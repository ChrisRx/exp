package env

import (
	"cmp"
	"fmt"
	"reflect"
	"unsafe"

	"go.chrisrx.dev/x/strings"
)

// Print prints a representation of a struct to standard output. It using the
// same rules as [Parser].
func Print(v any, opts ...ParserOption) {
	parser := NewParser(opts...)
	p := printer{
		DisableAutoPrefix: parser.DisableAutoPrefix,
		RootPrefix:        parser.RootPrefix,
		RequireTagged:     parser.RequireTagged,
		ptrs:              make(ptrmap),
	}
	if err := p.Print(v); err != nil {
		panic(err)
	}
}

type printer struct {
	DisableAutoPrefix bool
	RootPrefix        string
	RequireTagged     bool

	ptrs   ptrmap
	indent int
}

func (p *printer) Print(v any) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}
	if rv.Type().Name() != "" {
		fmt.Printf("%v\n", rv.Type().Name())
		p.indent++
	}
	field := Field{}
	if p.RootPrefix != "" {
		field.prefixes = append(field.prefixes, p.RootPrefix)
	}
	p.print(rv, field)
	return nil
}

func (p *printer) print(rv reflect.Value, field Field) {
	switch rv.Kind() {
	case reflect.Pointer:
		if _, ok := p.ptrs.get(rv); ok {
			return
		}
		p.ptrs.add(rv)
		p.print(reflect.Indirect(rv), field)
		p.ptrs.remove(rv)
	case reflect.Struct:
		prefixes := field.prefixes
		switch {
		case !p.DisableAutoPrefix && !field.Anonymous:
			prefixes = append(prefixes, cmp.Or(field.Env, strings.ToSnakeCase(field.Name)))
		default:
			prefixes = append(prefixes, field.Env)
		}
		for i := range rv.NumField() {
			p.indent++
			field := newField(rv.Type().Field(i), prefixes...)
			fmt.Print(strings.Repeat("  ", p.indent-1))
			fmt.Print(field.Name)
			fmt.Println("{")
			if field.Env != "" {
				fmt.Print(strings.Repeat("  ", p.indent))
				fmt.Printf("env=%s\n", field.Key())
			}
			if field.Default != "" {
				fmt.Print(strings.Repeat("  ", p.indent))
				fmt.Printf("default=%s\n", field.Default)
			}
			if rv.Type().Field(i).Tag.Get("layout") != "" {
				fmt.Print(strings.Repeat("  ", p.indent))
				fmt.Printf("layout=%s\n", field.Layout)
			}
			p.print(rv.Field(i), field)
			fmt.Print(strings.Repeat("  ", p.indent-1))
			fmt.Println("}")
			p.indent--
		}
	default:
		if rv.IsValid() && !rv.IsZero() {
			fmt.Print(strings.Repeat("  ", p.indent))
			fmt.Printf("value=%v\n", rv)
		}
	}
}

type ptrmap map[unsafe.Pointer]reflect.Value

func (p ptrmap) add(v reflect.Value) { p[unsafe.Pointer(v.Pointer())] = v.Elem() }
func (p ptrmap) get(v reflect.Value) (reflect.Value, bool) {
	v, ok := p[unsafe.Pointer(v.Pointer())]
	return v, ok
}
func (p ptrmap) remove(v reflect.Value) { delete(p, unsafe.Pointer(v.Pointer())) }
