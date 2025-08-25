package context_test

import (
	"fmt"

	"go.chrisrx.dev/x/context"
	"go.chrisrx.dev/x/strings"
)

type Auth struct {
	Headers map[string]string
}

func (a *Auth) String() string {
	return fmt.Sprintf(strings.Dedent(`
	Auth{
		%s
	}`), a.Headers)
}

func ExampleKey() {
	AuthKey := context.Key[*Auth]()

	ctx := AuthKey.WithValue(context.Background(), &Auth{
		Headers: map[string]string{
			"Authentication": "Bearer ...",
		},
	})

	fmt.Println(AuthKey.Has(ctx))
	fmt.Println(AuthKey.Value(ctx))
	// Output: true
	// Auth{
	// 	map[Authentication:Bearer ...]
	// }
}
