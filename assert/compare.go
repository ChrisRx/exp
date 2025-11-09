package assert

import (
	"cmp"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"go.chrisrx.dev/x/assert/internal/diff"
	"go.chrisrx.dev/x/assert/internal/slices"
)

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
		return cmp.Compare(rx.String(), ry.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(rx.Int(), ry.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmp.Compare(rx.Uint(), ry.Uint())
	case reflect.Float32, reflect.Float64:
		return cmp.Compare(rx.Float(), ry.Float())
	case reflect.String:
		return cmp.Compare(rx.String(), ry.String())
	default:
		panic(fmt.Errorf("compare received unhandled type: %T", x))
	}
}

func contains[S ~[]E, E any](s S, v E) bool {
	return slices.Contains(slices.Map(s, func(elem E) bool {
		return len(Diff(elem, v)) == 0
	}), true)
}

func isTime(v reflect.Value) bool {
	return v.Type().PkgPath() == "time" && v.Type().Name() == "Time"
}

func isDuration(v reflect.Value) bool {
	return v.Type().PkgPath() == "time" && v.Type().Name() == "Duration"
}

func isComparable(v any) bool {
	rv := reflect.ValueOf(v)
	if isTime(rv) {
		return true
	}
	return rv.Comparable()
}

func isZero(v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return reflect.Indirect(rv).IsZero()
}

func newMatcherFunc(s string) (func(string) bool, error) {
	if s == "" {
		return func(s string) bool {
			return s == ""
		}, nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		return re.MatchString(s)
	}, nil
}

func Diff[T any](expected, actual T) []byte {
	return diff.Diff([]byte(Sprint(expected)), []byte(Sprint(actual)))
}
