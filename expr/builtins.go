package expr

import "reflect"

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
	"Something": reflect.ValueOf(Something{}),
}

type Something struct {
	S string
	N int
}
