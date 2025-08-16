package pg

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host         string `env:"HOST" default:"localhost"`
	Port         int    `env:"PORT" default:"5432"`
	Username     string `env:"USERNAME" required:"false"`
	Password     string `env:"PASSWORD" required:"false"`
	DatabaseName string `env:"NAME"`

	// standard
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" default:"30s"`
	SSLMode        SSLMode       `env:"SSL_MODE" default:"prefer"`

	// pgx-specific
	MinPoolConns int `env:"MIN_POOL_CONNS" required:"false"`
	MaxPoolConns int `env:"MAX_POOL_CONNS" required:"false"`
}

func (c Config) String() string {
	v := make(url.Values)
	v.Set("sslmode", c.SSLMode.String())
	if c.ConnectTimeout != 0 {
		v.Set("connect_timeout", strconv.Itoa(int(c.ConnectTimeout.Seconds())))
	}
	if c.MinPoolConns != 0 {
		v.Set("min_pool_conns", strconv.Itoa(c.MinPoolConns))
	}
	if c.MaxPoolConns != 0 {
		v.Set("max_pool_conns", strconv.Itoa(c.MaxPoolConns))
	}

	u := &url.URL{
		Scheme:   "postgresql",
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.DatabaseName,
		RawQuery: v.Encode(),
	}
	switch {
	case c.Username != "" && c.Password != "":
		u.User = url.User(c.Username)
	case c.Username != "" && c.Password == "":
		u.User = url.UserPassword(c.Username, c.Password)
	}

	return u.String()
}

type SSLMode int

const (
	Prefer SSLMode = iota
	Disable
	Allow
	Require
	VerifyCA
	VerifyFull
)

func (s SSLMode) String() string {
	switch s {
	case Prefer:
		return "prefer"
	case Disable:
		return "disable"
	case Allow:
		return "allow"
	case Require:
		return "require"
	case VerifyCA:
		return "verify-ca"
	case VerifyFull:
		return "verify-full"
	default:
		return "prefer"
	}
}

func (s SSLMode) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *SSLMode) UnmarshalText(data []byte) error {
	switch strings.ToLower(string(data)) {
	case "prefer":
		*s = Prefer
	case "disable":
		*s = Disable
	case "allow":
		*s = Allow
	case "require":
		*s = Require
	case "verify-ca":
		*s = VerifyCA
	case "verify-full":
		*s = VerifyFull
	default:
		return fmt.Errorf("received invalid SSLMode: %q", data)
	}
	return nil
}
