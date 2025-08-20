package env

import (
	"cmp"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"go.chrisrx.dev/x/ptr"
)

type CustomParserFunc func(string, Field) (any, error)

var customParserFuncs = make(map[reflect.Type]CustomParserFunc)

func Register[T any](fn CustomParserFunc) {
	rt := reflect.TypeFor[T]()
	if rt.Kind() == reflect.Pointer {
		panic(fmt.Errorf("cannot register type %v: must not be pointer", rt))
	}
	customParserFuncs[rt] = func(s string, field Field) (any, error) {
		// avoid needing customer parsers to handle empty input
		s = cmp.Or(s, field.Default)
		if s == "" {
			return nil, nil
		}
		if isExpr(s) {
			rv, err := eval(s)
			if err != nil {
				return nil, err
			}
			if t, ok := typeAssert[time.Time](rv); ok {
				return t, nil
			}
		}
		return fn(s, field)
	}
}

func LookupFunc(rt reflect.Type) (CustomParserFunc, bool) {
	fn, ok := customParserFuncs[indirectType(rt)]
	return fn, ok
}

func init() {
	Register[time.Time](func(s string, field Field) (any, error) {
		return time.Parse(field.Layout, s)
	})

	Register[time.Duration](func(s string, field Field) (any, error) {
		return time.ParseDuration(s)
	})

	Register[url.URL](func(s string, field Field) (any, error) {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		return ptr.From(u), nil
	})

	Register[[]byte](func(s string, field Field) (any, error) {
		return []byte(s), nil
	})
}
