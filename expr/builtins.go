package expr

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
)

var isTesting atomic.Bool

func enableTesting() {
	isTesting.Store(true)
}

var testingTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// TODO(ChrisRx): generate the builtins/docs

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
	"some": reflect.ValueOf(func(v any) bool { return !ptr.IsZero(v) }),
	"none": reflect.ValueOf(func(v any) bool { return ptr.IsZero(v) }),

	// basic type casts
	"int": reflect.ValueOf(func(v any) int {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(rv.Uint())
		case reflect.Float32, reflect.Float64:
			return int(rv.Float())
		case reflect.String:
			return int(must.Get0(strconv.ParseInt(rv.String(), 10, 64)))
		default:
			return 0
		}
	}),
	"float": reflect.ValueOf(func(v any) float64 {
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
	"string": reflect.ValueOf(func(v any) string { return fmt.Sprint(v) }),

	// math
	"min": reflect.ValueOf(math.Min),
	"max": reflect.ValueOf(math.Max),
	"rand": reflect.ValueOf(func(args ...any) int {
		switch len(args) {
		case 1:
			return rand.IntN(take[int](args, 0))
		case 2:
			min := take[int](args, 1)
			max := take[int](args, 1)
			return rand.IntN(max-min) + min
		default:
			return rand.Int()
		}
	}),
	"random": reflect.ValueOf(rand.Float64),

	// strings
	"startswith": reflect.ValueOf(strings.HasPrefix),
	"endswith":   reflect.ValueOf(strings.HasSuffix),
	"trim":       reflect.ValueOf(strings.Trim),
	"upper":      reflect.ValueOf(strings.ToUpper),
	"lower":      reflect.ValueOf(strings.ToLower),
	"split":      reflect.ValueOf(strings.Split),
	"atoi":       reflect.ValueOf(func(s string) int { return must.Get0(strconv.Atoi(s)) }),
	"itoa":       reflect.ValueOf(strconv.Itoa),
	"quote":      reflect.ValueOf(strconv.Quote),
	"unquote":    reflect.ValueOf(func(s string) string { return must.Get0(strconv.Unquote(s)) }),

	// fmt
	"print":    reflect.ValueOf(fmt.Print),
	"printf":   reflect.ValueOf(fmt.Printf),
	"println":  reflect.ValueOf(fmt.Println),
	"sprint":   reflect.ValueOf(fmt.Sprint),
	"sprintf":  reflect.ValueOf(fmt.Sprintf),
	"sprintln": reflect.ValueOf(fmt.Sprintln),

	// os
	"getwd":    reflect.ValueOf(func() string { return must.Get0(os.Getwd()) }),
	"tempdir":  reflect.ValueOf(os.TempDir),
	"joinpath": reflect.ValueOf(filepath.Join),
	"getenv":   reflect.ValueOf(os.Getenv),

	// time
	"now": reflect.ValueOf(func() time.Time {
		if isTesting.Load() {
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

	// net
	"parse_mac": reflect.ValueOf(net.ParseMAC),
	"parse_ip":  reflect.ValueOf(net.ParseIP),
	"split_addr": reflect.ValueOf(func(s string) struct {
		Host string
		Port int
	} {
		type addr struct {
			Host string
			Port int
		}
		host, port, _ := net.SplitHostPort(s)
		i, _ := strconv.Atoi(port)
		return addr{Host: host, Port: i}
	}),

	// hash
	"hmac": reflect.ValueOf(func(args ...any) string {
		return fmt.Sprintf("%x", hmac.New(sha256.New, take[[]byte](args, 0)).Sum(take[[]byte](args, 1)))
	}),
	"md5": reflect.ValueOf(func(input any) string {
		return fmt.Sprintf("%x", md5.Sum(convert[[]byte](input)))
	}),
	"sha1": reflect.ValueOf(func(input any) string {
		return fmt.Sprintf("%x", sha1.Sum(convert[[]byte](input)))
	}),
	"sha256": reflect.ValueOf(func(input any) string {
		return fmt.Sprintf("%x", sha256.Sum256(convert[[]byte](input)))
	}),
}

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
		"json": {
			"Encode": reflect.ValueOf(func(v any) string {
				return string(must.Get0(json.Marshal(v)))
			}),
		},
		"base64": {
			"Decode": reflect.ValueOf(func(s string) string {
				data, _ := base64.StdEncoding.DecodeString(s)
				return string(data)
			}),
			"Encode": reflect.ValueOf(func(input any) string {
				switch input := input.(type) {
				case []byte:
					return base64.StdEncoding.EncodeToString(input)
				case string:
					return base64.StdEncoding.EncodeToString([]byte(input))
				default:
					return ""
				}
			}),
		},
	}
})

func convert[T any](v any) T {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeFor[T]()
	if rv.CanConvert(rt) {
		return rv.Convert(rt).Interface().(T)
	}
	var zero T
	return zero
}

func take[T any](elems []any, index int) T {
	var zero T
	return takeOr(elems, index, zero)
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
