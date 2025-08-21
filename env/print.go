package env

import (
	"cmp"
	"fmt"
	"reflect"
	"unsafe"

	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/strings"
)

func Print(v any, opts ...ParserOption) error {
	parser := NewParser(opts...)
	p := printer{
		DisableAutoPrefix: parser.DisableAutoPrefix,
		Namespace:         parser.Namespace,
		RequireTagged:     parser.RequireTagged,
		ptrs:              make(ptrmap),
	}
	return p.Print(v)
}

type printer struct {
	DisableAutoPrefix bool
	Namespace         string
	RequireTagged     bool

	ptrs   ptrmap
	indent int
}

func (p *printer) Print(v any) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("must provide a struct pointer, received %T", v)
	}
	for i := range rv.NumField() {
		field := NewField(rv, i)
		fmt.Println(field.Name)
		p.print(rv.Field(i), field)
	}
	return nil
}

func (p *printer) print(rv reflect.Value, field Field) {
	switch rv.Kind() {
	case reflect.Pointer:
		if _, ok := p.ptrs.get(rv); ok {
			return
		}
		p.ptrs.add(rv)
		defer p.ptrs.remove(rv)
		p.print(reflect.Indirect(rv), field)
	case reflect.Struct:
		prefixes := field.prefixes
		switch {
		case !p.DisableAutoPrefix && !field.Anonymous:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace, strings.ToSnakeCase(field.Name)))
		default:
			prefixes = append(prefixes, cmp.Or(field.Env, field.Namespace))
		}
		var maxFieldNameLen int
		if rv.NumField() > 0 {
			maxFieldNameLen = slices.Max(slices.Map(slices.N(rv.NumField()), func(i int) int {
				name := rv.Type().Field(i).Name
				return len(name)
			}))
		}
		for i := range rv.NumField() {
			p.indent++
			field := NewField(rv, i, prefixes...)
			fmt.Print(strings.Repeat("  ", p.indent))
			fmt.Print(field.Name)
			fmt.Print(strings.Repeat(" ", max(maxFieldNameLen-len(field.Name)+1, 0)))
			if field.Env != "" {
				fmt.Print(field.Key())
			}
			fmt.Println()
			if field.Default != "" {
				fmt.Print(strings.Repeat("  ", p.indent+1))
				fmt.Printf("default=%s\n", field.Default)
			}
			if rv.Type().Field(i).Tag.Get("layout") != "" {
				fmt.Print(strings.Repeat("  ", p.indent+1))
				fmt.Printf("layout=%s\n", field.Layout)
			}
			p.print(rv.Field(i), field)
			p.indent--
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
