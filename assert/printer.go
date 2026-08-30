package assert

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"go.chrisrx.dev/x/assert/internal/slices"
	"go.chrisrx.dev/x/internal/reflectx"
)

func Fprint(w io.Writer, v any) {
	p := &printer{
		indent: "        ",
		ptrs: ptrset{
			m: make(map[unsafe.Pointer]reflect.Value),
		},
		w: w,
	}
	p.do(reflect.ValueOf(v))
}

func Print(v any) {
	Fprint(os.Stdout, v)
}

func Sprint(v any) string {
	var b bytes.Buffer
	Fprint(&b, v)
	return b.String()
}

type printer struct {
	indent string
	ptrs   ptrset
	depth  int
	w      io.Writer
}

func (p *printer) fprint(a ...any) {
	fmt.Fprint(p.w, a...)
}

func (p *printer) fprintf(format string, a ...any) {
	fmt.Fprintf(p.w, format, a...)
}

func (p *printer) do(rv reflect.Value) {
	if !rv.IsValid() {
		p.fprint(p.w, "(invalid)(nil)")
		return
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			if rv.Kind() == reflect.Pointer {
				p.ptrs.add(rv)
			}
			typ := replaceAnyType(rv.Type().String())
			typ = strings.ReplaceAll(typ, "[]uint8", "[]byte")
			p.fprintf("(%v)(nil)", typ)
			return
		}
	}

	// well-known
	switch rv.Kind() {
	case reflect.Int64:
		if isDuration(rv) {
			p.fprint(time.Duration(rv.Int()))
			return
		}
	case reflect.Struct:
		if isTime(rv) {
			p.fprint(rv.Interface().(time.Time).Format(time.RFC3339Nano))
			return
		}
	case reflect.Array, reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			p.fprintf("[]byte(%q)", rv.Bytes())
			return
		}
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.fprint(strconv.FormatInt(rv.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p.fprint(strconv.FormatUint(rv.Uint(), 10))
	case reflect.Float32:
		p.fprint(strconv.FormatFloat(rv.Float(), 'f', -1, 32))
	case reflect.Float64:
		p.fprint(strconv.FormatFloat(rv.Float(), 'f', -1, 64))
	case reflect.String:
		p.fprintf(`"%v"`, rv.Interface())
	case reflect.Map:
		if v, ok := p.ptrs.get(rv); ok {
			p.doPointer(rv, v)
			return
		}
		p.ptrs.m[unsafe.Pointer(rv.Pointer())] = rv
		defer p.ptrs.remove(rv)
		keys := rv.MapKeys()
		if len(keys) == 0 {
			p.fprintf("%s{}", replaceAnyType(rv.Type().String()))
			return
		}
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
		})

		p.fprint(replaceAnyType(rv.Type().String()))
		p.fprint("{\n")
		p.depth++
		for _, k := range keys {
			p.fprint(strings.Repeat(p.indent, p.depth))
			p.do(k)
			p.fprint(": ")
			p.do(rv.MapIndex(k))
			p.fprint(",\n")
		}
		p.depth--
		p.fprint(strings.Repeat(p.indent, p.depth))
		p.fprint("}")
	case reflect.Pointer:
		if v, ok := p.ptrs.get(rv); ok {
			p.doPointer(rv, v)
			return
		}
		p.ptrs.add(rv)
		defer p.ptrs.remove(rv)
		p.fprint("*")
		elem := rv.Elem()
		if elem.Kind() == reflect.Struct {
			p.doStruct(elem, isProtoMessage(rv))
			return
		}
		p.do(elem)
	case reflect.Struct:
		p.doStruct(rv, false)
	case reflect.Array, reflect.Slice:
		p.fprint(fmt.Sprintf("%s{\n", replaceAnyType(rv.Type().String())))
		p.depth++
		for i := range rv.Len() {
			p.fprint(strings.Repeat(p.indent, p.depth))
			p.do(rv.Index(i))
			p.fprint("\n")
		}
		p.depth--
		p.fprint(strings.Repeat(p.indent, p.depth))
		p.fprint("}")
	case reflect.Interface:
		elem := rv.Elem()
		p.do(elem)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		p.fprintf("(%v)(%#x)", rv.Type(), rv.Pointer())
	case reflect.Uintptr:
		p.fprintf("(%v)(%#x)", rv.Type(), rv.Uint())
	default:
		if !rv.CanInterface() {
			p.fprintf("%v", rv.String())
			return
		}
		p.fprint(replaceAnyType(fmt.Sprintf("%v", rv.Interface())))
	}
}

func (p *printer) doStruct(rv reflect.Value, exportedOnly bool) {
	p.fprint(replaceAnyType(rv.Type().String()))
	p.fprint("{\n")
	var maxFieldNameLen int
	if rv.NumField() > 0 {
		maxFieldNameLen = slices.Max(slices.Map(slices.N(rv.NumField()), func(i int) int {
			return len(rv.Type().Field(i).Name)
		}))
	}
	p.depth++
	for i := range rv.NumField() {
		ft := rv.Type().Field(i)
		fv := rv.Field(i)
		if exportedOnly && !ft.IsExported() {
			continue
		}
		if !fv.CanInterface() && fv.CanAddr() {
			fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
		}
		if !fv.CanAddr() {
			fv = reflectx.MakeAddressableField(rv, i)
		}
		p.fprint(strings.Repeat(p.indent, p.depth))
		p.fprint(ft.Name)
		p.fprint(": ")
		p.fprint(strings.Repeat(" ", max(maxFieldNameLen-len(ft.Name), 0)))
		p.do(fv)
		p.fprint("\n")
	}
	p.depth--
	p.fprint(strings.Repeat(p.indent, p.depth))
	p.fprint("}")
}

// isProtoMessage reports whether rv implements proto.Message, checked by
// method name rather than a real interface assertion so that this package
// doesn't need to import google.golang.org/protobuf.
func isProtoMessage(rv reflect.Value) bool {
	t := rv.Type()
	if t.Kind() != reflect.Pointer {
		t = reflect.PointerTo(t)
	}
	m, ok := t.MethodByName("ProtoReflect")
	return ok && m.Type.NumIn() == 1 && m.Type.NumOut() == 1
}

func (p *printer) doPointer(ptr, v reflect.Value) {
	if !ptr.CanAddr() {
		switch ptr.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.String, reflect.UnsafePointer:
			p.fprintf("(*%v)(%v)", replaceAnyType(v.Type().String()), ptr.UnsafePointer())
			return
		default:
			p.fprintf("(*%v)(<unknown>)", replaceAnyType(v.Type().String()))
			return
		}
	}
	p.fprintf("(*%v)(%v)", replaceAnyType(v.Type().String()), ptr.Addr())
}

type ptrset struct {
	m map[unsafe.Pointer]reflect.Value
}

func (p *ptrset) add(v reflect.Value) {
	p.m[unsafe.Pointer(v.Pointer())] = v.Elem()
}

func (p *ptrset) get(v reflect.Value) (reflect.Value, bool) {
	ptr, ok := p.m[unsafe.Pointer(v.Pointer())]
	return ptr, ok
}

func (p *ptrset) remove(v reflect.Value) {
	delete(p.m, unsafe.Pointer(v.Pointer()))
}

func replaceAnyType(s string) string {
	return strings.ReplaceAll(s, "interface {}", "any")
}
