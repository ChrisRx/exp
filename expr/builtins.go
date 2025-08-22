package expr

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"go.chrisrx.dev/x/must"
)

var isTesting bool
var testingTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

var builtins = map[string]reflect.Value{
	"len": reflect.ValueOf(func(v any) int {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer {
			rv = reflect.Indirect(rv)
		}
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.String:
			return rv.Len()
		default:
			return 0
		}
	}),

	// basic type casts
	"int": reflect.ValueOf(func(v any) any {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(rv.Uint())
		case reflect.Float32, reflect.Float64:
			return int(rv.Float())
		default:
			return 0
		}
	}),
	"float": reflect.ValueOf(func(v any) any {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint())
		case reflect.Float32, reflect.Float64:
			return rv.Float()
		default:
			return float64(0)
		}
	}),
	"string": reflect.ValueOf(func(v any) string {
		return fmt.Sprint(v)
	}),

	// math
	"min": reflect.ValueOf(math.Min),
	"max": reflect.ValueOf(math.Max),

	// strings
	"startswith": reflect.ValueOf(strings.HasPrefix),
	"endswith":   reflect.ValueOf(strings.HasSuffix),
	"trim":       reflect.ValueOf(strings.Trim),
	"upper":      reflect.ValueOf(strings.ToUpper),
	"lower":      reflect.ValueOf(strings.ToLower),

	// fmt
	"print": reflect.ValueOf(func(args ...any) {
		fmt.Print(args...)
	}),
	"printf": reflect.ValueOf(func(format string, args ...any) {
		fmt.Printf(format, args...)
	}),
	"println": reflect.ValueOf(func(args ...any) {
		fmt.Println(args...)
	}),
	"sprint": reflect.ValueOf(func(args ...any) string {
		return fmt.Sprint(args...)
	}),
	"sprintf": reflect.ValueOf(func(format string, args ...any) string {
		return fmt.Sprintf(format, args...)
	}),
	"sprintln": reflect.ValueOf(func(args ...any) string {
		return fmt.Sprintln(args...)
	}),

	// time
	"now": reflect.ValueOf(func() time.Time {
		if isTesting {
			return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		return time.Now()
	}),
	"date": reflect.ValueOf(func(args ...any) time.Time {
		return time.Date(
			take[int](args, 0),            // year
			takeOr(args, 1, time.January), // month
			take[int](args, 2),            // day
			take[int](args, 3),            // hour
			take[int](args, 4),            // min
			take[int](args, 5),            // sec
			take[int](args, 6),            // nsec
			takeOr(args, 7, time.UTC),     // loc
		)
	}),
	"duration": reflect.ValueOf(func(s string) time.Duration {
		return must.Get0(time.ParseDuration(s))
	}),
}

func take[T any](elems []any, index int) T {
	var zero T
	if len(elems)-1 < index {
		return zero
	}
	rv := reflect.ValueOf(elems[index])
	rt := reflect.TypeFor[T]()
	if rv.CanConvert(rt) {
		return rv.Convert(rt).Interface().(T)
	}
	if t, ok := elems[index].(T); ok {
		return t
	}
	return zero
}

func takeOr[T any](elems []any, index int, orElse T) T {
	if len(elems)-1 < index {
		return orElse
	}
	rv := reflect.ValueOf(elems[index])
	rt := reflect.TypeFor[T]()
	if rv.CanConvert(rt) {
		return rv.Convert(rt).Interface().(T)
	}
	if t, ok := elems[index].(T); ok {
		return t
	}
	return orElse
}
