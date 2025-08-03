package assert

import (
	"cmp"
	"fmt"
	"hash/maphash"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"
)

var seed = maphash.MakeSeed()

func hash(v any) uint64 {
	h := new(maphash.Hash)
	h.SetSeed(seed)
	_, _ = fmt.Fprint(h, v)
	return h.Sum64()
}

func equal(a, b any, opts ...Option) bool {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return hash(a) == hash(b)
}

// assert is going to be used in pretty much every package so it needs to copy
// code that might exist in those packages already to prevent an import cycle.
func Map[T any, R any](col []T, fn func(elem T) R) []R {
	results := make([]R, len(col))
	for i, v := range col {
		results[i] = fn(v)
	}
	return results
}

func contains[S ~[]E, E any](s S, v E) bool {
	return slices.Contains(Map(s, func(elem E) uint64 {
		return hash(elem)
	}), hash(v))
}

func print(v any) string {
	rv := reflect.Indirect(reflect.ValueOf(v))

	var sb strings.Builder
	switch rv.Kind() {
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
	// 	reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
	// 	reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
	// 	reflect.Complex64, reflect.Complex128:
	// case reflect.Chan, reflect.Map:
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

func Diff(a, b any) string {
	ra := reflect.ValueOf(a)
	rb := reflect.ValueOf(b)

	var sb strings.Builder
	sb.WriteString("expected:\n\t")
	if ra.Kind() == reflect.Struct {
		sb.WriteString(ra.Type().String())
	}
	fmt.Fprint(&sb, a)
	sb.WriteString("\n")
	sb.WriteString("actual:\n\t")
	if ra.Kind() == reflect.Struct {
		sb.WriteString(rb.Type().String())
	}
	fmt.Fprint(&sb, b)
	return sb.String()
}

func compare(x, y any) int {
	rx := reflect.ValueOf(x)
	ry := reflect.ValueOf(y)

	if rx.Type() != ry.Type() {
		panic(fmt.Errorf("compare: cannot compare unlike types: %T <> %T", x, y))
	}

	switch rx.Kind() {
	case reflect.Struct:
		if isTime(rx) {
			return time.Time.Compare(rx.Interface().(time.Time), ry.Interface().(time.Time))
		}
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(rx.Int(), ry.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmp.Compare(rx.Uint(), ry.Uint())
	case reflect.Float32, reflect.Float64:
		return cmp.Compare(rx.Float(), ry.Float())
	default:
		panic(fmt.Errorf("compare received unhandled type: %T", x))
	}
}

func isTime(v reflect.Value) bool {
	return v.Type().PkgPath() == "time" && v.Type().Name() == "Time"
}

func isComparable(v any) bool {
	rv := reflect.ValueOf(v)
	if isTime(rv) {
		return true
	}
	return rv.Comparable()
}

func newMatcher(s string) (func(string) bool, error) {
	if s == "" {
		return func(s string) bool {
			return s == ""
		}, nil
	}
	var sb strings.Builder
	if s[0] != '^' {
		sb.WriteString("^")
	}
	sb.WriteString(s)
	if s[len(s)-1] != '$' {
		sb.WriteString("$")
	}
	re, err := regexp.Compile(sb.String())
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		return re.MatchString(s)
	}, nil
}
