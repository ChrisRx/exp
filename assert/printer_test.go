package assert_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"unsafe"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/assert/internal/diff"
)

func TestPrinter(t *testing.T) {
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

func TestPrintTypes(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{
			name:  "chan",
			value: make(chan error),
		},
		{
			name:  "chan",
			value: (chan error)(nil),
		},
		{
			name:  "uintptr",
			value: (uintptr)(0x12f25ac9a230),
		},
		{
			name:  "unsafe.Pointer",
			value: unsafe.Pointer(uintptr(0x12f25ac9a230)),
		},
		{
			name:  "any",
			value: (any)("something"),
		},
		{
			name: "struct",
			value: struct {
				v          any
				unexported int
				Exported   string
			}{
				v:          "value",
				unexported: 5,
				Exported:   "exported",
			},
		},
		{
			name:  "bytes",
			value: []byte("byte value"),
		},
		{
			name:  "time.Duration",
			value: 15 * time.Second,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Print(tt.value)
		})
	}
}

func TestDo(t *testing.T) {
	a := map[string]any{}
	a["circular"] = map[string]any{
		"a": a,
	}
	b := map[string]any{
		"a": a,
		"b": a,
	}
	assert.Print(b)

	t.Run("slice", func(t *testing.T) {
		s := "test"
		a := []any{&s}
		a = append(a, a)
		a = append(a, a)
		b := map[string]any{
			"a": a,
			"b": a,
		}
		assert.Print(b)
	})
}
