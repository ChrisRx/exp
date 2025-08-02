package env

import (
	"github.com/caarlos0/env/v11"
)

// MustParse parses tags for the provided struct and sets values from
// environment variables. It will only set values that have the `env` tag, and
// if the tag is present it must be provided, otherwise this panics.
func MustParse(v any) {
	if err := env.ParseWithOptions(v, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}
}
