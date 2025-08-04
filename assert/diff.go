package assert

import (
	"fmt"
	"reflect"
	"strings"
)

func print(v any) string {
	rv := reflect.Indirect(reflect.ValueOf(v))

	var sb strings.Builder
	switch rv.Kind() {
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
	// 	reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
	// 	reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
	// 	reflect.Complex64, reflect.Complex128:
	// case reflect.Map:
	// case reflect.Chan:
	// case reflect.Ptr:
	// case reflect.Func:
	case reflect.Struct:
		sb.WriteString(rv.Type().String())
		sb.WriteString("{\n")
		for i := range rv.NumField() {
			ft := rv.Type().Field(i)
			fv := rv.Field(i)
			if fv.CanInterface() {
				sb.WriteString(fmt.Sprintf("\t%s: %v,\n", ft.Name, fv.Interface()))
			}
		}
		sb.WriteString("}")
		return sb.String()
	case reflect.Array, reflect.Slice:
		sb.WriteString(fmt.Sprintf("[]%s{", rv.Type().String()))
		for i := range rv.Len() {
			sb.WriteString(print(rv.Index(i).Interface()))
			sb.WriteString("\n")
		}
		sb.WriteString("}")
		return sb.String()
	default:
		return fmt.Sprint(v)
	}
}
