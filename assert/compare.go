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

func equal(a, b any, opts ...Option) bool {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return hash(a) == hash(b)
}

var seed = maphash.MakeSeed()

func hash(v any) uint64 {
	h := new(maphash.Hash)
	h.SetSeed(seed)
	_, _ = fmt.Fprint(h, v)
	return h.Sum64()
}

func contains[S ~[]E, E any](s S, v E) bool {
	return slices.Contains(imap(s, func(elem E) uint64 {
		return hash(elem)
	}), hash(v))
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
