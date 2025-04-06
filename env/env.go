package env

import (
	"github.com/caarlos0/env/v11"
	"github.com/samber/lo"
)

// MustParse parses tags for the provided struct and sets values from
// environment variables. It will only set values that have the `env` tag, and
// if the tag is present it must be provided, otherwise this panics.
func MustParse(v any) {
	lo.Must0(env.ParseWithOptions(v, env.Options{
		RequiredIfNoDef: true,
	}))
}
