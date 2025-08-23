package expr

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

// TODO(ChrisRx): maybe rethink this
var packages = sync.OnceValue(func() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"fmt": {
			"Print":   reflect.ValueOf(fmt.Print),
			"Printf":  reflect.ValueOf(fmt.Printf),
			"Println": reflect.ValueOf(fmt.Println),
			"Sprint":  reflect.ValueOf(fmt.Sprint),
			"Sprintf": reflect.ValueOf(fmt.Sprintf),
		},
		"math": {
			"Abs":   reflect.ValueOf(math.Abs),
			"Acos":  reflect.ValueOf(math.Acos),
			"Asin":  reflect.ValueOf(math.Asin),
			"Atan":  reflect.ValueOf(math.Atan),
			"Ceil":  reflect.ValueOf(math.Ceil),
			"Cos":   reflect.ValueOf(math.Cos),
			"Exp":   reflect.ValueOf(math.Exp),
			"Log":   reflect.ValueOf(math.Log),
			"Max":   reflect.ValueOf(math.Max),
			"Min":   reflect.ValueOf(math.Min),
			"Round": reflect.ValueOf(math.Round),
			"Sin":   reflect.ValueOf(math.Sin),
			"Tan":   reflect.ValueOf(math.Tan),
		},
		"time": {
			"Time":        reflect.ValueOf(time.Time{}),
			"Local":       reflect.ValueOf(time.Local),
			"UTC":         reflect.ValueOf(time.UTC),
			"Date":        reflect.ValueOf(time.Date),
			"Now":         reflect.ValueOf(time.Now),
			"Nanosecond":  reflect.ValueOf(time.Nanosecond),
			"Millisecond": reflect.ValueOf(time.Millisecond),
			"Second":      reflect.ValueOf(time.Second),
			"Minute":      reflect.ValueOf(time.Minute),
			"Hour":        reflect.ValueOf(time.Hour),
			"Duration":    reflect.ValueOf(time.Duration(0)),
		},
		"net": {
			"ParseIP":   reflect.ValueOf(net.ParseIP),
			"ParseCIDR": reflect.ValueOf(net.ParseCIDR),
		},
		"strings": {
			"HasPrefix": reflect.ValueOf(strings.HasPrefix),
			"HasSuffix": reflect.ValueOf(strings.HasSuffix),
		},
	}
})
