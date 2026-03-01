package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/strings"
	"go.chrisrx.dev/x/structs"
)

// Field represents a parsed struct field.
type Field struct {
	structs.Field

	// tags
	Env      string
	Validate string
	Required bool

	DeferredMethod string

	prefixes []string
}

func newField(st reflect.StructField, prefixes ...string) Field {
	return Field{
		Field:          structs.Field(st),
		Env:            st.Tag.Get("env"),
		Validate:       st.Tag.Get("validate"),
		Required:       must.Get0(strconv.ParseBool(st.Tag.Get("required"))),
		DeferredMethod: st.Tag.Get("method"),
		prefixes:       slices.FilterMap(prefixes, strings.ToUpper),
	}
}

// Key returns the environment variable name in screaming snake case. It
// includes any prefixes defined for this field.
func (f Field) Key() string {
	return strings.Join(append(f.prefixes, f.Env), "_")
}

func (f Field) set(rv reflect.Value) error {
	s, ok := os.LookupEnv(f.Key())
	if !ok {
		ok, err := f.SetDefault(rv)
		if err != nil {
			return err
		}
		if !ok && f.Required {
			return fmt.Errorf("required field not set: %v", f.Key())
		}
		return nil
	}
	return structs.ParseField(s, rv)
}
