package expr

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
)

type Something struct {
	S string
	N int
}

func init() {
	enableTesting()

	builtins["Something"] = reflect.ValueOf(Something{})
}

type testCase struct {
	name     string
	input    string
	env      map[string]reflect.Value
	expected any
}

func (tc testCase) run(t *testing.T) {
	t.Helper()
	v, err := Eval(tc.input, Env(tc.env))
	if err != nil {
		t.Fatal(err)
	}
	if !v.CanInterface() {
		t.Fatalf("returned value cannot interface")
	}
	assert.Equal(t, tc.expected, v.Interface(), tc.name)
}

func runAll(t *testing.T, cases []testCase) {
	t.Helper()
	for i, tc := range cases {
		switch {
		case tc.name != "":
			t.Run(tc.name, tc.run)
		default:
			tc.name = fmt.Sprintf("test #%d", i+1)
			tc.run(t)
		}
	}
}

func TestEval(t *testing.T) {
	t.Run("arithmetic", func(t *testing.T) {
		runAll(t, []testCase{
			{name: "add", input: `12+11`, expected: 12 + 11},
			{name: "subtract", input: `25-14`, expected: 25 - 14},
			{name: "multiply", input: `10*1024`, expected: 10 * 1024},
			{name: "divide", input: `100/10`, expected: 100 / 10},
			{name: "right shift", input: `32 >> 1`, expected: 32 >> 1},
			{name: "left shift", input: `1 << 32`, expected: 1 << 32},
			{name: "right shift", input: `32 >> 1`, expected: 32 >> 1},
		})
	})

	t.Run("integer literals", func(t *testing.T) {
		runAll(t, []testCase{
			{name: "", input: "42", expected: 42},
			{name: "", input: "4_2", expected: 4_2},
			{name: "", input: "8600", expected: 8600},
			{name: "", input: "0_600", expected: 0_600},
			{name: "", input: "0o600", expected: 0o600},
			{name: "", input: "0O600", expected: 0o600},
			{name: "", input: "0xBadFace", expected: 0xBadFace},
			{name: "", input: "0xBad_Face", expected: 0xBad_Face},
			{name: "", input: "0x_67_7a_2f_cc_40_c6", expected: 0x_67_7a_2f_cc_40_c6},
			// TODO: This truncates large numbers into a float64 but not sure if this
			// should do this or should just error.
			{name: "", input: "170141183460469231731687303715884105727", expected: float64(170141183460469231731687303715884105727)},
			{name: "", input: "170_141183_460469_231731_687303_715884_105727", expected: float64(170141183460469231731687303715884105727)},
		})
	})

	t.Run("floating-point literals", func(t *testing.T) {
		runAll(t, []testCase{
			{name: "", input: "0.", expected: 0},
			{name: "", input: "72.40", expected: 72.40},
			{name: "", input: "072.40", expected: 072.40},
			{name: "", input: "2.71828", expected: 2.71828},
			{name: "", input: "1.e+0", expected: 1.e+0},
			{name: "", input: "6.67428e-11", expected: 6.67428e-11},
			{name: "", input: "1E6", expected: 1e6},
			{name: "", input: ".25", expected: .25},
			{name: "", input: ".12345E+5", expected: .12345e+5},
			{name: "", input: "1_5.", expected: 1_5.},
			{name: "", input: "0.15e+0_2", expected: 0.15e+0_2},
			{name: "", input: "0x1p-2", expected: 0x1p-2},
			{name: "", input: "0x2.p10", expected: 0x2.p10},
			{name: "", input: "0x1.Fp+0", expected: 0x1.Fp+0},
			{name: "", input: "0X.8p-0", expected: 0x.8p-0},
			{name: "", input: "0X_1FFFP-16", expected: 0x_1FFFp-16},
			{name: "", input: "0x15e-2", expected: 0x15e - 2},
		})
	})

	t.Run("operators", func(t *testing.T) {
		runAll(t, []testCase{
			{name: "equals", input: `1 == 1`, expected: true},
			{name: "less than", input: `len("hello") < 30`, expected: true},
			{name: "greater than", input: `len("hello") > 4`, expected: true},
			{name: "less than equals", input: `len("hello") <= 5`, expected: true},
			{name: "greater than equals", input: `len("hello") >= 5`, expected: true},
		})
	})

	t.Run("time", func(t *testing.T) {
		runAll(t, []testCase{
			{
				input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)",
				expected: testingTime,
			},
			{
				input:    "time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Minute)",
				expected: testingTime.Add(-1 * time.Minute),
			},
			{input: `time.Time{}`, expected: time.Time{}},
			{input: `time.Duration(5)`, expected: time.Duration(5)},
			{input: `date(2020, 1, 1).add(duration("-1m"))`, expected: testingTime.Add(-1 * time.Minute)},
			{input: `now().add(duration("-1m"))`, expected: testingTime.Add(-1 * time.Minute)},
			{input: `now() + duration("-1m")`, expected: testingTime.Add(-1 * time.Minute)},
			{input: `now().is_zero()`, expected: false},
		})
	})

	t.Run("builtins", func(t *testing.T) {
		runAll(t, []testCase{
			{
				input:    `fmt.Sprintf("%s: %d", "count", 100)`,
				expected: fmt.Sprintf("%s: %d", "count", 100),
			},
			{
				input:    `sprintf("%s: %d", "count", 100)`,
				expected: fmt.Sprintf("%s: %d", "count", 100),
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
			{
				name:     "json",
				input:    `json.encode(Something{S:"testing", N: 5})`,
				expected: `{"S":"testing","N":5}`,
			},
		})
	})

	t.Run("misc", func(t *testing.T) {
		runAll(t, []testCase{
			{
				name:     "struct literal method call",
				input:    `time.Time{}.is_zero()`,
				expected: true,
			},
			{
				name:     "struct literal method call with parens",
				input:    `(time.Time{}).is_zero()`,
				expected: true,
			},
			{
				name:     "composite literal",
				input:    `Something{S: "testing", N: 5}`,
				expected: Something{S: "testing", N: 5},
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
		})
	})
}
