package env_test

import (
	"testing"

	"go.chrisrx.dev/x/env"
	"go.chrisrx.dev/x/env/testdata/pg"
)

func TestPrint(t *testing.T) {
	err := env.Print(struct {
		Database pg.Config
		Nested   struct {
			Database pg.Config
		}
	}{})
	if err != nil {
		t.Fatal(err)
	}
}
