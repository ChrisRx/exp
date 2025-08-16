package env_test

import (
	"log/slog"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/env/testdata/log"
	"go.chrisrx.dev/x/env/testdata/pubsub"
	"go.chrisrx.dev/x/env/testdata/spanner"
)

func TestEnv(t *testing.T) {
	assert.WithEnviron(t, map[string]string{
		"USERS_SPANNER_PROJECT":     "test-project",
		"USERS_SPANNER_INSTANCE":    "test-instance",
		"USERS_SPANNER_DATABASE":    "test-database",
		"TASKS_PUBSUB_TOPIC":        "test-pubsub-topic",
		"TASKS_PUBSUB_SUBSCRIPTION": "test-pubsub-subscription",
	}, func() {
		expected := struct {
			Tasks pubsub.Config
			Users spanner.Config
		}{
			Tasks: pubsub.Config{
				Topic:        "test-pubsub-topic",
				Subscription: "test-pubsub-subscription",
			},
			Users: spanner.Config{
				Project:  "test-project",
				Instance: "test-instance",
				Database: "test-database",
			},
		}

		t.Run("anonymous struct", func(t *testing.T) {
			var opts = env.MustParseAs[struct {
				Tasks pubsub.Config
				Users spanner.Config
			}]()

			assert.Equal(t, expected, opts)
		})

		t.Run("anonymous struct pointer", func(t *testing.T) {
			var opts struct {
				Tasks pubsub.Config
				Users spanner.Config
			}
			env.MustParse(&opts)

			assert.Equal(t, expected, opts)
		})
	})
}

func TestParse(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		environ  map[string]string
		expected any
		err      error
	}{
		{
			name: "slice of strings",
			input: new(struct {
				StringSlice []string `env:"STRING_SLICE"`
			}),
			environ: map[string]string{
				"STRING_SLICE": "a,b,c",
			},
			expected: &struct {
				StringSlice []string `env:"STRING_SLICE"`
			}{
				StringSlice: []string{"a", "b", "c"},
			},
		},
		{
			name: "auto tag structs",
			input: new(struct {
				*log.Options
				Tasks pubsub.Config
				Users spanner.Config
			}),
			environ: map[string]string{
				"LOG_LEVEL":                 "DEBUG",
				"LOG_FORMAT":                "json",
				"LOG_ADD_SOURCE":            "true",
				"USERS_SPANNER_PROJECT":     "test-project",
				"USERS_SPANNER_INSTANCE":    "test-instance",
				"USERS_SPANNER_DATABASE":    "test-database",
				"TASKS_PUBSUB_TOPIC":        "test-pubsub-topic",
				"TASKS_PUBSUB_SUBSCRIPTION": "test-pubsub-subscription",
			},
			expected: &struct {
				*log.Options
				Tasks pubsub.Config
				Users spanner.Config
			}{
				Options: &log.Options{
					Level: func() *slog.LevelVar {
						lvl := new(slog.LevelVar)
						lvl.Set(slog.LevelDebug)
						return lvl
					}(),
					Format:    log.JSONFormat,
					AddSource: true,
				},
				Tasks: pubsub.Config{
					Topic:        "test-pubsub-topic",
					Subscription: "test-pubsub-subscription",
				},
				Users: spanner.Config{
					Project:  "test-project",
					Instance: "test-instance",
					Database: "test-database",
				},
			},
		},
		{
			name: "default values",
			input: new(struct {
				DefaultString   string        `env:"DEFAULT_STRING" default:"default string"`
				DefaultDuration time.Duration `env:"DEFAULT_DURATION" default:"10m"`
			}),
			expected: &struct {
				DefaultString   string        `env:"DEFAULT_STRING" default:"default string"`
				DefaultDuration time.Duration `env:"DEFAULT_DURATION" default:"10m"`
			}{
				DefaultString:   "default string",
				DefaultDuration: 10 * time.Minute,
			},
		},
		{
			name: "time",
			input: new(struct {
				Time time.Time `env:"TIME"`
			}),
			environ: map[string]string{
				"TIME": "2020-12-30T01:30:55Z",
			},
			expected: &struct {
				Time time.Time `env:"TIME"`
			}{
				Time: time.Date(2020, 12, 30, 1, 30, 55, 0, time.UTC),
			},
		},
		{
			name:  "log options",
			input: new(log.Options),
			environ: map[string]string{
				"LOG_LEVEL":      "DEBUG",
				"LOG_FORMAT":     "json",
				"LOG_ADD_SOURCE": "true",
			},
			expected: &log.Options{
				Level: func() *slog.LevelVar {
					lvl := new(slog.LevelVar)
					lvl.Set(slog.LevelDebug)
					return lvl
				}(),
				Format:    log.JSONFormat,
				AddSource: true,
			},
		},
	}

	for _, tc := range cases {
		assert.WithEnviron(t, tc.environ, func() {
			err := env.Parse(tc.input)
			assert.Error(t, tc.err, err, tc.name)
			assert.Equal(t, tc.expected, tc.input, tc.name)
		})
	}
}
