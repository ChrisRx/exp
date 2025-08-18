package spanner

import "fmt"

type Config struct {
	Project  string `env:"SPANNER_PROJECT"`
	Instance string `env:"SPANNER_INSTANCE"`
	Database string `env:"SPANNER_DATABASE"`
}

func (c Config) String() string {
	return fmt.Sprintf("projects/%s/instances/%s/databases/%s", c.Project, c.Instance, c.Database)
}
