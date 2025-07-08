package ptr_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/ptr"
)

func TestIsZero(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "string - zero",
			input:    "",
			expected: true,
		},
		{
			name:     "string - non-zero",
			input:    "some value",
			expected: false,
		},
		{
			name:     "slice - zero (nil)",
			input:    ([]string)(nil),
			expected: true,
		},
		{
			name:     "slice - zero",
			input:    []string{},
			expected: false,
		},
		{
			name:     "slice - non-zero",
			input:    []string{"some value"},
			expected: false,
		},
		{
			name:     "function - zero",
			input:    (func())(nil),
			expected: true,
		},
		{
			name:     "function - non-zero",
			input:    func() {},
			expected: false,
		},
	}
	for _, tt := range cases {
		result := ptr.IsZero(tt.input)
		assert.EqualValues(t, tt.expected, result, tt.name)
	}
}

func ExampleNullable() {
	var s string
	fmt.Printf("(%[1]T)(%[1]v)\n", ptr.Nullable(&s))

	// Output: (*string)(<nil>)
}

func ExampleToNullable() {
	var s struct {
		Value *string
	}
	s.Value = ptr.ToNullable("")
	fmt.Println(s.Value)
	s.Value = ptr.ToNullable("not zero")
	fmt.Println(ptr.From(s.Value))

	// Output: <nil>
	// not zero
}
