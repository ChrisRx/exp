package convert

import (
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/ptr"
)

func TestConvert(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			input:    "2026-01-01T00:00:00Z",
			expected: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	t.Run("elem to elem", func(t *testing.T) {
		conversions = make(map[key]func(any, ...Option) (any, error))
		Register(func(s string, opts ...Option) (time.Time, error) {
			return time.Parse(time.RFC3339Nano, s)
		})
		for _, tt := range cases {
			v, err := Into[time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			v, err = Into[time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			p, err := Into[*time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
			p, err = Into[*time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
		}
	})

	t.Run("elem to ptr", func(t *testing.T) {
		conversions = make(map[key]func(any, ...Option) (any, error))
		Register(func(s string, opts ...Option) (*time.Time, error) {
			t, err := time.Parse(time.RFC3339Nano, s)
			if err != nil {
				return nil, err
			}
			return ptr.To(t), nil
		})
		for _, tt := range cases {
			v, err := Into[time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			v, err = Into[time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			p, err := Into[*time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
			p, err = Into[*time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
		}
	})

	t.Run("ptr to ptr", func(t *testing.T) {
		conversions = make(map[key]func(any, ...Option) (any, error))
		Register(func(s *string, opts ...Option) (*time.Time, error) {
			t, err := time.Parse(time.RFC3339Nano, ptr.From(s))
			if err != nil {
				return nil, err
			}
			return ptr.To(t), nil
		})
		for _, tt := range cases {
			v, err := Into[time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			v, err = Into[time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			p, err := Into[*time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
			p, err = Into[*time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
		}
	})

	t.Run("ptr to elem", func(t *testing.T) {
		conversions = make(map[key]func(any, ...Option) (any, error))
		Register(func(s *string, opts ...Option) (time.Time, error) {
			return time.Parse(time.RFC3339Nano, ptr.From(s))
		})
		for _, tt := range cases {
			v, err := Into[time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			v, err = Into[time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, v, tt.name)
			p, err := Into[*time.Time](tt.input)
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
			p, err = Into[*time.Time](ptr.To(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, ptr.To(tt.expected), p, tt.name)
		}
	})
}
