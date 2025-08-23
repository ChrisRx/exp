package env_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/env/testdata/log"
	"go.chrisrx.dev/x/env/testdata/pg"
	"go.chrisrx.dev/x/env/testdata/pubsub"
	"go.chrisrx.dev/x/env/testdata/spanner"
	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/ptr"
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
			var opts = env.MustParseFor[struct {
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
	t.Run("embedded structs", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"LOG_LEVEL":      "DEBUG",
			"LOG_FORMAT":     "json",
			"LOG_ADD_SOURCE": "true",
		}, func() {
			assert.Equal(t, struct{ *log.Options }{
				Options: &log.Options{
					Level: func() *slog.LevelVar {
						lvl := new(slog.LevelVar)
						lvl.Set(slog.LevelDebug)
						return lvl
					}(),
					Format:    log.JSONFormat,
					AddSource: true,
				},
			}, env.MustParseFor[struct{ *log.Options }](),
			)

			assert.Equal(t, struct{ log.Options }{
				Options: log.Options{
					Level: func() *slog.LevelVar {
						lvl := new(slog.LevelVar)
						lvl.Set(slog.LevelDebug)
						return lvl
					}(),
					Format:    log.JSONFormat,
					AddSource: true,
				},
			}, env.MustParseFor[struct{ log.Options }](),
			)
		})
	})

	t.Run("nested structs", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"USERS_SPANNER_PROJECT":          "test-project",
			"USERS_SPANNER_INSTANCE":         "test-instance",
			"USERS_SPANNER_DATABASE":         "test-database",
			"TASKS_PUBSUB_TOPIC":             "test-pubsub-topic",
			"TASKS_PUBSUB_SUBSCRIPTION":      "test-pubsub-subscription",
			"SELF_USERS_SPANNER_PROJECT":     "test-project-nested",
			"SELF_USERS_SPANNER_INSTANCE":    "test-instance-nested",
			"SELF_USERS_SPANNER_DATABASE":    "test-database-nested",
			"SELF_TASKS_PUBSUB_TOPIC":        "test-pubsub-topic-nested",
			"SELF_TASKS_PUBSUB_SUBSCRIPTION": "test-pubsub-subscription-nested",
		}, func() {
			type s struct {
				Tasks pubsub.Config
				Users spanner.Config
			}
			assert.Equal(t, s{
				Tasks: pubsub.Config{
					Topic:        "test-pubsub-topic",
					Subscription: "test-pubsub-subscription",
				},
				Users: spanner.Config{
					Project:  "test-project",
					Instance: "test-instance",
					Database: "test-database",
				},
			}, env.MustParseFor[s]())

			type s2 struct {
				Tasks *pubsub.Config
				Users spanner.Config
				Self  s
			}
			assert.Equal(t, &s2{
				Tasks: &pubsub.Config{
					Topic:        "test-pubsub-topic",
					Subscription: "test-pubsub-subscription",
				},
				Users: spanner.Config{
					Project:  "test-project",
					Instance: "test-instance",
					Database: "test-database",
				},
				Self: s{
					Tasks: pubsub.Config{
						Topic:        "test-pubsub-topic-nested",
						Subscription: "test-pubsub-subscription-nested",
					},
					Users: spanner.Config{
						Project:  "test-project-nested",
						Instance: "test-instance-nested",
						Database: "test-database-nested",
					},
				},
			}, env.MustParseFor[*s2]())
		})
	})

	t.Run("default values", func(t *testing.T) {
		type s struct {
			String   string        `env:"DEFAULT_STRING" default:"default string"`
			Duration time.Duration `env:"DEFAULT_DURATION" default:"10m"`
			Time     time.Time     `env:"DEFAULT_TIME" default:"2020-12-30" layout:"2006-01-02"`
		}
		assert.Equal(t, s{
			String:   "default string",
			Duration: 10 * time.Minute,
			Time:     time.Date(2020, 12, 30, 0, 0, 0, 0, time.UTC),
		}, env.MustParseFor[s]())

		var opts = env.MustParseFor[struct {
			Time    time.Time `env:"DEFAULT_TIME" $default:"now()"`
			String  string    `env:"DEFAULT_STRING" $default:"now().format('2006-01-02')"`
			String2 string    `env:"DEFAULT_STRING" $default:"now()" layout:"2006-01-02"`
			Number  int       `env:"DEFAULT_NUMBER" $default:"5 + 6"`
			IP      net.IP    `env:"IP_ADDRESS" $default:"net.ParseIP('127.0.0.1')"`
		}]()
		assert.WithinDuration(t, time.Now(), opts.Time, 10*time.Millisecond)
		assert.Equal(t, opts.String, opts.String2)
		assert.Equal(t, 11, opts.Number)
		assert.Equal(t, opts.IP, net.ParseIP("127.0.0.1"))
	})

	t.Run("slices", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"STRING_SLICE":          "a,b,c",
			"INT_SLICE":             "-1,0,1",
			"INT32_SLICE":           "1,2,3",
			"UINT_SLICE":            "1,2,3",
			"STRING_POINTER_SLICE":  "a,b,c",
			"INVALID_POINTER_SLICE": "<invalid>",
		}, func() {
			type s struct {
				StringSlice        []string  `env:"STRING_SLICE"`
				IntSlice           []int     `env:"INT_SLICE"`
				Int32Slice         []int32   `env:"INT32_SLICE"`
				UintSlice          []uint    `env:"UINT_SLICE"`
				StringPointerSlice []*string `env:"STRING_POINTER_SLICE"`
			}
			assert.Equal(t, s{
				StringSlice:        []string{"a", "b", "c"},
				IntSlice:           []int{-1, 0, 1},
				Int32Slice:         []int32{1, 2, 3},
				UintSlice:          []uint{1, 2, 3},
				StringPointerSlice: []*string{ptr.To("a"), ptr.To("b"), ptr.To("c")},
			}, env.MustParseFor[s]())

			assert.Error(t, "received unhandled value:.*", must.Get1(env.ParseFor[struct {
				S []*time.Location `env:"INVALID_POINTER_SLICE"`
			}]()))
		})
	})
	t.Run("bytes", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"BYTES": "hello",
		}, func() {
			opts := env.MustParseFor[struct {
				Bytes []byte `env:"BYTES"`
			}]()
			assert.Equal(t, []byte("hello"), opts.Bytes)
		})
	})

	t.Run("maps", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"COOKIES":          "key1=value1,key2=value2,key3=value3",
			"NUMBERS":          "key1=1,key2=2,key3=3",
			"OOPS_ALL_NUMBERS": "1=1,2=2,3=3",
			"DURATION":         "key1=10s,key2=20s,key3=30s",
		}, func() {
			opts := env.MustParseFor[struct {
				Cookies     map[string]string         `env:"COOKIES"`
				Numbers     map[string]int            `env:"NUMBERS"`
				AllNumbers  map[int]int               `env:"OOPS_ALL_NUMBERS"`
				Duration    map[string]time.Duration  `env:"DURATION"`
				DurationPtr map[string]*time.Duration `env:"DURATION"`
			}]()
			assert.Equal(t, map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}, opts.Cookies)
			assert.Equal(t, map[string]int{
				"key1": 1,
				"key2": 2,
				"key3": 3,
			}, opts.Numbers)
			assert.Equal(t, map[int]int{
				1: 1,
				2: 2,
				3: 3,
			}, opts.AllNumbers)
			assert.Equal(t, map[string]time.Duration{
				"key1": 10 * time.Second,
				"key2": 20 * time.Second,
				"key3": 30 * time.Second,
			}, opts.Duration)
			assert.Equal(t, map[string]*time.Duration{
				"key1": ptr.To(10 * time.Second),
				"key2": ptr.To(20 * time.Second),
				"key3": ptr.To(30 * time.Second),
			}, opts.DurationPtr)
		})
	})

	t.Run("complex", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"DATABASE_HOST": "127.0.0.1",
			"DATABASE_PORT": "5432",
			"DATABASE_NAME": "postgres",
		}, func() {
			opts := env.MustParseFor[struct {
				Database pg.Config
			}]()

			assert.Equal(t, pg.Config{
				Host:           "127.0.0.1",
				Port:           5432,
				DatabaseName:   "postgres",
				ConnectTimeout: 30 * time.Second,
				SSLMode:        pg.Prefer,
			}, opts.Database)
			assert.Equal(t,
				"postgresql://127.0.0.1:5432/postgres?connect_timeout=30&sslmode=prefer",
				opts.Database.String(),
			)
		})

		assert.WithEnviron(t, map[string]string{
			"USERS_DB_NAME":            "users",
			"USERS_DB_SSL_MODE":        "verify-ca",
			"USERS_DB_CONNECT_TIMEOUT": "1m",
			"USERS_DB_MAX_POOL_CONNS":  "100",
		}, func() {
			opts := env.MustParseFor[struct {
				Database pg.Config `namespace:"USERS_DB"`
			}]()

			assert.Equal(t, pg.Config{
				Host:           "localhost",
				Port:           5432,
				DatabaseName:   "users",
				ConnectTimeout: 1 * time.Minute,
				SSLMode:        pg.VerifyCA,
				MaxPoolConns:   100,
			}, opts.Database)
			assert.Equal(t,
				"postgresql://localhost:5432/users?connect_timeout=60&max_pool_conns=100&sslmode=verify-ca",
				opts.Database.String(),
			)
		})
	})

	t.Run("custom parser funcs", func(t *testing.T) {
		type CustomType struct {
			S string
		}
		assert.Panic(t, fmt.Errorf("cannot register type %T: must not be pointer", &CustomType{}), func() {
			env.Register[*CustomType](func(field env.Field, s string) (any, error) {
				return CustomType{S: s}, nil
			})
		})
		env.Register[CustomType](func(field env.Field, s string) (any, error) {
			return CustomType{S: s}, nil
		})
		priv, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		pub, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		assert.WithEnviron(t, map[string]string{
			"DURATION":     "10s",
			"DURATION_PTR": "90s",
			"URL":          "https://www.google.com",
			"CUSTOM_TYPE":  "hi",
			"RSA_PUBKEY":   string(pem.EncodeToMemory(&pem.Block{Bytes: pub})),
			"IP_ADDRESS":   "127.0.0.1",
		}, func() {
			opts := env.MustParseFor[struct {
				Duration      time.Duration  `env:"DURATION"`
				DurationPtr   *time.Duration `env:"DURATION_PTR"`
				URL           url.URL        `env:"URL"`
				URLPtr        *url.URL       `env:"URL"`
				CustomType    CustomType     `env:"CUSTOM_TYPE"`
				CustomTypePtr *CustomType    `env:"CUSTOM_TYPE"`
				PubKey        *rsa.PublicKey `env:"RSA_PUBKEY"`
				IP            net.IP         `env:"IP_ADDRESS" validate:"some(parse_ip('127.0.0.1'))"`
			}]()

			assert.Equal(t, 10*time.Second, opts.Duration)
			assert.Equal(t, ptr.To(90*time.Second), opts.DurationPtr)
			assert.Equal(t, url.URL{Scheme: "https", Host: "www.google.com"}, opts.URL)
			assert.Equal(t, &url.URL{Scheme: "https", Host: "www.google.com"}, opts.URLPtr)
			assert.Equal(t, 10*time.Second, opts.Duration)
			assert.Equal(t, CustomType{S: "hi"}, opts.CustomType)
			assert.Equal(t, &CustomType{S: "hi"}, opts.CustomTypePtr)
			assert.Equal(t, &priv.PublicKey, opts.PubKey)
			assert.Equal(t, opts.IP, net.ParseIP("127.0.0.1"))
		})
	})

	t.Run("expressions", func(t *testing.T) {
		opts := env.MustParseFor[struct {
			Result0 string `env:"RESULT" $default:"fmt.Sprint(math.Round(math.Cos(45)*180))"`
			Result1 string `env:"RESULT" $default:"sprint(math.Round(math.Cos(45)*180))"`
			Result2 string `env:"RESULT" $default:"sprint(min(1,2))"`
		}]()
		assert.Equal(t, "95", opts.Result0)
		assert.Equal(t, "95", opts.Result1)
		assert.Equal(t, "1", opts.Result2)
	})

	t.Run("validate", func(t *testing.T) {
		assert.WithEnviron(t, map[string]string{
			"UUID": "7b2b2c53-0d22-44af-8ac5-e080434352b2",
		}, func() {
			assert.NoError(t, must.Get1(env.ParseFor[struct {
				ID string `env:"UUID" validate:"len(ID) == 36"`
			}]()), "validate length successful")

			assert.Error(t, "field ID failed validation", must.Get1(env.ParseFor[struct {
				ID string `env:"UUID" validate:"len(ID) == 37"`
			}]()), "validate length failed")

			assert.NoError(t, must.Get1(env.ParseFor[struct {
				ID      string `env:"UUID" validate:"some(ID)"`
				OtherID string `env:"UUID" validate:"!none(OtherID)"`
			}]()), "value not zero")

			assert.Error(t, "undefined: WrongField", must.Get1(env.ParseFor[struct {
				OtherID string `env:"UUID" validate:"none(WrongField)"`
			}]()), "wrong field symbol")

			assert.NoError(t, must.Get1(env.ParseFor[struct {
				Num int `env:"NUM" $default:"min(200,150)" validate:"Num < 200 && Num > 100"`
			}]()), "cannot validate number range")
		})
	})
}
