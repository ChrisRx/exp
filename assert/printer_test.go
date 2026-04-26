package assert_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/assert/internal/diff"
	"go.chrisrx.dev/x/assert/internal/testdata/test3"
)

func TestPrinter(t *testing.T) {
	t.Run("proto", func(t *testing.T) {
		assert.Equal(t,
			&test3.TestAllTypes{
				SingularInt32: 1,
				OptionalInt64: new(int64(1)),
				SingularBytes: []byte("test\x00\x0a"),
			},
			&test3.TestAllTypes{
				SingularInt32: 1,
				OptionalInt64: new(int64(1)),
				SingularBytes: []byte("test\x00\x0a"),
			},
		)
	})

	t.Run("empty map", func(t *testing.T) {
		assert.Equal(t,
			"(map[string]string)(nil)",
			assert.Sprint(*new(map[string]string)),
		)
		assert.Equal(t,
			"map[string]string{}",
			assert.Sprint(make(map[string]string)),
		)
	})

	t.Run("print", func(t *testing.T) {
		t.Skip()
		type Nested struct {
			String string
			T      time.Time
		}
		type Embedded struct {
			IsEmbedded bool
		}
		type S struct {
			FloatValue float64
			Duration   time.Duration
			Chan       chan error
			Any        any
			Map        map[string]any
			Time       time.Time
			Nested     Nested
			Embedded
			NestedPtr *S
			Self      *S
			Func      func(ctx context.Context) string

			private time.Duration
			t       time.Time
		}

		s := &S{
			Duration: 100 * time.Millisecond,
			Any:      "something",
			Nested: Nested{
				String: "idk",
				T:      time.Now(),
			},
			Embedded: Embedded{
				IsEmbedded: true,
			},
			NestedPtr: &S{
				FloatValue: 0.12345,
				Any:        "idk",
			},
			Map: map[string]any{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			private: 1 * time.Hour,
			t:       time.Now(),
		}
		s.Self = s
		assert.Print(s)

		t.Run("diff", func(t *testing.T) {
			t.Skip()
			s2 := &S{
				Duration: 100 * time.Millisecond,
				Any:      "something",
				Nested: Nested{
					String: "idk",
					T:      time.Now(),
				},
				Embedded: Embedded{
					IsEmbedded: true,
				},
				NestedPtr: &S{
					FloatValue: 0.12345,
					Any:        "idk",
				},
				Map: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
				private: 1 * time.Hour,
				t:       time.Now(),
			}
			d := diff.Diff([]byte(assert.Sprint(s)), []byte(assert.Sprint(s2)))
			fmt.Printf("%s\n", d)
		})
	})
}
