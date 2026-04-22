package assert

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert/internal/diff"
)

func TestPrinter(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		Equal(t, "(map[string]string)(nil)", Sprint(*new(map[string]string)))
		Equal(t, "map[string]string{}", Sprint(make(map[string]string)))
	})

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
	Print(s)

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
	d := diff.Diff([]byte(Sprint(s)), []byte(Sprint(s2)))
	fmt.Printf("%s\n", d)
}
