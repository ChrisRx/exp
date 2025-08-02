package assert

import (
	"fmt"
	"hash/maphash"
	"reflect"
	"slices"
	"strings"
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
