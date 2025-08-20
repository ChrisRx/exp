package expr_test

import (
	"fmt"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/expr"
)

func TestEval(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected any
	}{
		{
			input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)",
			expected: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Minute)",
			expected: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Minute),
		},
		{
			input:    `fmt.Sprintf("%s: %d", "count", 100)`,
			expected: fmt.Sprintf("%s: %d", "count", 100),
		},
		{
			input:    `1 == 1`,
			expected: true,
		},
		{
			input:    `len("hello") >= 5 && len("hello") < 30`,
			expected: true,
		},
		{
			input:    `time.Time{}`,
			expected: time.Time{},
		},
		{
			input:    `time.Duration(5)`,
			expected: time.Duration(5),
		},
		{
			input:    `1 << 32`,
			expected: 1 << 32,
		},
		{
			input:    `32 >> 1`,
			expected: 32 >> 1,
		},
		{
			input:    `Something{S: "testing", N: 5}`,
			expected: expr.Something{S: "testing", N: 5},
		},
	}

	for _, tc := range cases {
		v, err := expr.Eval(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if !v.CanInterface() {
			t.Fatalf("returned value cannot interface")
		}

		assert.Equal(t, tc.expected, v.Interface())
	}
}
