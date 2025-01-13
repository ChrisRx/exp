package ptr_test

import (
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
