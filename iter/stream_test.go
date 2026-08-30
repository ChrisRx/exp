package iter_test

import (
	"cmp"
	"net/http"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/iter"
	"go.chrisrx.dev/x/maps"
	"go.chrisrx.dev/x/slices"
)

func TestStream(t *testing.T) {
	t.Run("map/reduce", func(t *testing.T) {
		assert.Equal(t,
			"abcdefghijklmnopqrstuvwxyz",
			iter.N(26).Map(func(i int) string {
				return string(rune('a' + i))
			}).Reduce(func(acc, c string) string {
				return acc + c
			}),
		)
		assert.Equal(t,
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			iter.N(26).Map(func(i int) rune {
				return rune('A' + i)
			}).Fold(func(acc string, c rune) string {
				var sb strings.Builder
				sb.WriteString(acc)
				sb.WriteRune(c)
				return sb.String()
			}),
		)
		assert.Equal(t,
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			string(iter.N(26).
				Map(func(i int) rune {
					return rune('a' + i)
				}).
				Map(unicode.ToUpper).
				Fold(utf8.AppendRune),
			),
		)
	})

	t.Run("uniq", func(t *testing.T) {
		assert.Equal(t,
			[]int{1, 2, 3, 4, 5},
			iter.From(slices.Values(
				[]int{1, 1, 1, 2, 2, 3, 4, 4, 5, 5, 5, 5},
			)).Uniq().Collect(),
		)
		assert.Equal(t,
			[]int{1, 5, 2, 4, 3},
			iter.From(slices.Values(
				[]int{1, 5, 1, 2, 4, 4, 3, 4, 5, 1, 5, 2},
			)).Uniq().Collect(),
			"out-of-order",
		)
	})

	t.Run("map", func(t *testing.T) {
		assert.Equal(t,
			"abcdefghijklmnopqrstuvwxyz",
			string(iter.Map(slices.Values(slices.N(26)), func(i int) rune {
				return rune('a' + i)
			}).Fold(utf8.AppendRune)),
		)
	})

	t.Run("skip", func(t *testing.T) {
		assert.Equal(t,
			"fghijklmnopqrstuvwxyz",
			string(iter.N(26).
				Skip(5).
				Map(func(i int) rune {
					return rune('a' + i)
				}).
				Fold(utf8.AppendRune),
			),
		)
	})

	t.Run("sort", func(t *testing.T) {
		assert.Equal(t,
			[]int{5, 4, 3, 2, 1},
			iter.From(slices.Values(
				[]int{1, 3, 2, 5, 4},
			)).SortFunc(func(a, b int) int {
				return -cmp.Compare(a, b)
			}),
		)
		assert.Equal(t,
			[]int{5, 4, 3, 2, 1},
			iter.From(slices.Values(
				[]int{1, 1, 2, 2, 3, 4, 4, 5, 5, 5, 5},
			)).Uniq().SortFunc(func(a, b int) int {
				return -cmp.Compare(a, b)
			}),
			"sort uniq",
		)
	})

	t.Run("find", func(t *testing.T) {
		letter := iter.N(26).
			Map(func(i int) string {
				return string(rune('a' + i))
			}).
			Find(func(c string) bool {
				return c == "y"
			})
		assert.Equal(t, "y", letter)

		stream := iter.Of(maps.Entries(http.Header{
			"Accept": []string{"application/json"},
			"Cookie": []string{"k1=v1,k2=v2"},
		})).
			Filter(func(kv maps.Entry[string, []string]) bool {
				return kv.Key == "Cookie"
			})

		values := stream.
			FlatMap(func(kv maps.Entry[string, []string]) []maps.Entry[string, string] {
				return slices.FlatMap(kv.Value, func(value string) []maps.Entry[string, string] {
					return slices.FilterMap(strings.Split(value, ","), func(v string) maps.Entry[string, string] {
						parts := strings.SplitN(v, "=", 2)
						if len(parts) != 2 {
							return maps.Entry[string, string]{}
						}
						return maps.KV(parts[0], parts[1])
					})
				})
			})
		assert.Equal(t, []maps.Entry[string, string]{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
		}, values.Collect())
		assert.Equal(t, maps.Entry[string, string]{
			Key:   "k2",
			Value: "v2",
		}, values.Find(func(kv maps.Entry[string, string]) bool {
			return kv.Key == "k2"
		}))

		values = stream.
			FlatMap(func(kv maps.Entry[string, []string]) []maps.Entry[string, string] {
				return iter.Of(kv.Value).FlatMap(func(s string) []string {
					return strings.Split(s, ",")
				}).Map(func(v string) []string {
					return strings.SplitN(v, "=", 2)
				}).Filter(func(v []string) bool {
					return len(v) == 2
				}).Map(func(v []string) maps.Entry[string, string] {
					return maps.KV(v[0], v[1])
				}).Collect()
			})
		assert.Equal(t, []maps.Entry[string, string]{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
		}, values.Collect())
		assert.Equal(t, maps.Entry[string, string]{
			Key:   "k2",
			Value: "v2",
		}, values.Find(func(kv maps.Entry[string, string]) bool {
			return kv.Key == "k2"
		}))

		values = stream.
			FlatMap(func(kv maps.Entry[string, []string]) []maps.Entry[string, string] {
				var results []maps.Entry[string, string]
				for _, v := range kv.Value {
					for v := range strings.SplitSeq(v, ",") {
						parts := strings.SplitN(v, "=", 2)
						if len(parts) != 2 {
							continue
						}
						results = append(results, maps.KV(parts[0], parts[1]))
					}
				}
				return results
			})
		assert.Equal(t, []maps.Entry[string, string]{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
		}, values.Collect())
		assert.Equal(t, maps.Entry[string, string]{
			Key:   "k2",
			Value: "v2",
		}, values.Find(func(kv maps.Entry[string, string]) bool {
			return kv.Key == "k2"
		}))
	})
}
