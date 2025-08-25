package expr

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

func init() {
	enableTesting()
}

type Something struct {
	S string
	N int
}

func TestEval(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		env      map[string]reflect.Value
		expected any
	}{
		{
			input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)",
			expected: testingTime,
		},
		{
			input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Minute)",
			expected: testingTime.Add(-1 * time.Minute),
		},
		{
			input:    `fmt.Sprintf("%s: %d", "count", 100)`,
			expected: fmt.Sprintf("%s: %d", "count", 100),
		},
		{
			input:    `sprintf("%s: %d", "count", 100)`,
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
			expected: Something{S: "testing", N: 5},
		},
		{
			input:    `date(2020, 1, 1).add(duration("-1m"))`,
			expected: testingTime.Add(-1 * time.Minute),
		},
		{
			input:    `now().add(duration("-1m"))`,
			expected: testingTime.Add(-1 * time.Minute),
		},
		{
			input:    `now() + duration("-1m")`,
			expected: testingTime.Add(-1 * time.Minute),
		},
		{
			input:    `now().is_zero()`,
			expected: false,
		},
		{
			input:    `time.Time{}.is_zero()`,
			expected: true,
		},
		{
			input:    `(time.Time{}).is_zero()`,
			expected: true,
		},
		{
			name:  "non-privileged port",
			input: `split_addr(self).port > 1024`,
			env: map[string]reflect.Value{
				"self": reflect.ValueOf(":8080"),
			},
			expected: true,
		},
		{
			name:  "implicit argument passing",
			input: `split_addr().port > 1024`,
			env: map[string]reflect.Value{
				"self": reflect.ValueOf(":8080"),
			},
			expected: true,
		},
		{
			input:    `hmac("some key", "data")`,
			expected: "6461746170d19601c3b9123566d9f6cec756de0d8fb3e57d2474dd82c3005edee7106ad7",
		},
		{
			input:    `md5("data")`,
			expected: "8d777f385d3dfec8815d20f7496026dc",
		},
		{
			input:    `base64.encode("data")`,
			expected: "ZGF0YQ==",
		},
		{
			input:    `base64.decode("ZGF0YQ==")`,
			expected: "data",
		},
	}

	builtins["Something"] = reflect.ValueOf(Something{})

	for _, tc := range cases {
		v, err := Eval(tc.input, Env(tc.env))
		if err != nil {
			t.Fatal(err)
		}
		if !v.CanInterface() {
			t.Fatalf("returned value cannot interface")
		}

		assert.Equal(t, tc.expected, v.Interface())
	}
}
