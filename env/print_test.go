package env_test

import (
	"testing"

	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/env/testdata/pg"
)

func TestPrint(t *testing.T) {
	env.Print(struct {
		Database pg.Config
		Nested   struct {
			Database pg.Config
		}
	}{})
}
