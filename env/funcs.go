package env

import (
	"reflect"
	"time"
)

type CustomParserFunc func(string, FieldTags) (any, error)

var customParserFuncs = make(map[reflect.Type]CustomParserFunc)

func RegisterCustomParserFunc[T any](fn CustomParserFunc) {
	customParserFuncs[reflect.TypeFor[T]()] = fn
}

func init() {
	RegisterCustomParserFunc[time.Time](func(s string, tags FieldTags) (any, error) {
		return time.Parse(tags.Layout, s)
	})

	RegisterCustomParserFunc[time.Duration](func(s string, tags FieldTags) (any, error) {
		return time.ParseDuration(s)
	})
}
