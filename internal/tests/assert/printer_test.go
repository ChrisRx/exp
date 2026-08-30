package assert_test

import (
	"strings"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/internal/tests/assert/testdata/test3"
)

func TestProto(t *testing.T) {
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
	t.Run("proto skip", func(t *testing.T) {
		msg := &test3.TestAllTypes_NestedMessage{
			A: 123,
			Corecursive: &test3.TestAllTypes{
				SingularInt32: 1,
				OptionalInt64: new(int64(1)),
				SingularBytes: []byte("test\x00\x0a"),
			},
		}
		assert.Equal(t,
			false,
			strings.Contains(assert.Sprint(msg), "impl.MessageState{"),
		)
	})
}
