package spanner

type Config struct {
	Project  string `env:"SPANNER_PROJECT"`
	Instance string `env:"SPANNER_INSTANCE"`
	Database string `env:"SPANNER_DATABASE"`
}
