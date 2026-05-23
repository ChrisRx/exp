package uuid_test

import (
	"strings"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/uuid"
)

func TestUUID(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty input", input: "", expected: "6c62272e-07bb-8142-a2b8-21756295c58d"},
		{name: "small input", input: "", expected: "6c62272e-07bb-8142-a2b8-21756295c58d"},
		{name: "large input", input: strings.Repeat("a", 8*1024*1024), expected: "9a3dfde3-1047-829c-ab11-3a3e4815c58d"},
	}
	for _, tt := range cases {
		assert.Equal(t, tt.expected, uuid.NewV8(tt.input), tt.name)
	}
}
