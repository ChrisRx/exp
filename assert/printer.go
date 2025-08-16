package assert

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"go.chrisrx.dev/x/assert/internal/slices"
)

func Sprint(v any) string {
	p := &printer{
		indent: "        ",
		ptrs: ptrset{
			m: make(map[unsafe.Pointer]reflect.Value),
		},
	}
	return p.getIndent(0) + p.sprint(reflect.ValueOf(v), 0)
}

func Print(v any) {
	fmt.Println(Sprint(v))
}

type printer struct {
	indent string
	ptrs   ptrset
}

func (p *printer) getIndent(depth int) string {
	return strings.Repeat(p.indent, depth)
}

func (p *printer) sprint(rv reflect.Value, depth int) string {
	if !rv.IsValid() {
		return "(invalid)(nil)"
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return fmt.Sprintf("(%v)(nil)", replaceAnyType(rv.Type().String()))
		}
	}

	switch rv.Kind() {
	case reflect.Int64:
		if isDuration(rv) {
			return fmt.Sprint(time.Duration(rv.Int()))
		}
	case reflect.Struct:
		if isTime(rv) {
			return rv.Interface().(time.Time).Format(time.RFC3339Nano)
		}
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isDuration(rv) {
			return fmt.Sprint(time.Duration(rv.Int()))
		}
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.String:
		return replaceAnyType(fmt.Sprintf(`"%v"`, rv.Interface()))
	case reflect.Map:
		keys := rv.MapKeys()
		maxKeyLen := slices.Max(slices.Map(keys, func(key reflect.Value) int {
			return len(key.String())
		}))
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
		})

		var sb strings.Builder
		sb.WriteString(replaceAnyType(rv.Type().String()))
		sb.WriteString("{\n")
		for _, k := range keys {
			key := p.sprint(k, depth)
			value := rv.MapIndex(k)
			sb.WriteString(p.getIndent(depth + 1))
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(strings.Repeat(" ", max(maxKeyLen-len(key), 0)))
			sb.WriteString(p.sprint(value, depth))
			sb.WriteString(",\n")
		}
		sb.WriteString(p.getIndent(depth))
		sb.WriteString("}")
		return sb.String()
	case reflect.Chan:
		return rv.Type().String()
	case reflect.Pointer:
		if v, ok := p.ptrs.get(rv); ok {
			return fmt.Sprintf("(*%v)(%v)", replaceAnyType(v.Type().String()), rv.Addr())
		}
		p.ptrs.add(rv)
		defer p.ptrs.remove(rv)
		return "*" + p.sprint(rv.Elem(), depth)
	case reflect.Func:
		return rv.Type().String()
	case reflect.Struct:
		var sb strings.Builder
		sb.WriteString(rv.Type().String())
		sb.WriteString("{\n")
		var maxFieldNameLen int
		if rv.NumField() > 0 {
			maxFieldNameLen = slices.Max(slices.Map(slices.N(rv.NumField()), func(i int) int {
				name := rv.Type().Field(i).Name
				return len(name)
			}))
		}
		for i := range rv.NumField() {
			ft := rv.Type().Field(i)
			fv := rv.Field(i)
			if !fv.CanInterface() && fv.CanAddr() {
				fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
			}
			if !fv.CanAddr() {
				fv = makeAddressableField(rv, i)
			}
			sb.WriteString(p.getIndent(depth + 1))
			sb.WriteString(ft.Name)
			sb.WriteString(": ")
			sb.WriteString(strings.Repeat(" ", max(maxFieldNameLen-len(ft.Name), 0)))
			sb.WriteString(p.sprint(fv, depth+1))
			sb.WriteString("\n")
		}
		sb.WriteString(p.getIndent(depth))
		sb.WriteString("}")
		return sb.String()
	case reflect.Array, reflect.Slice:
		var sb strings.Builder
		sb.WriteString(p.getIndent(depth))
		sb.WriteString(fmt.Sprintf("[]%s{\n", rv.Type().String()))
		for i := range rv.Len() {
			sb.WriteString(p.getIndent(depth + 1))
			sb.WriteString(p.sprint(rv.Index(i), depth+1))
			sb.WriteString("\n")
		}
		sb.WriteString(p.getIndent(depth))
		sb.WriteString("}")
		return p.getIndent(depth) + sb.String()
	case reflect.Interface:
		return p.sprint(rv.Elem(), depth)
	default:
		if !rv.CanInterface() {
			rv = reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem()
		}
		return replaceAnyType(fmt.Sprintf("%v", rv.Interface()))
	}
}

type ptrset struct {
	m map[unsafe.Pointer]reflect.Value
}

func (p *ptrset) add(v reflect.Value) {
	p.m[unsafe.Pointer(v.Pointer())] = v.Elem()
}

func (p *ptrset) get(v reflect.Value) (reflect.Value, bool) {
	v, ok := p.m[unsafe.Pointer(v.Pointer())]
	return v, ok
}

func (p *ptrset) remove(v reflect.Value) {
	delete(p.m, unsafe.Pointer(v.Pointer()))
}

func replaceAnyType(s string) string {
	return strings.ReplaceAll(s, "interface {}", "any")
}

func makeAddressableField(rv reflect.Value, fieldNum int) reflect.Value {
	rv2 := reflect.New(rv.Type()).Elem()
	rv2.Set(rv)
	fv := rv2.Field(fieldNum)
	return reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
}
