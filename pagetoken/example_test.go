package pagetoken_test

import (
	"fmt"
	"time"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/pagetoken"
)

func ExampleParse() {
	token, err := pagetoken.Parse[pagetoken.Cursor[int]](
		pagetoken.Cursor[int]{After: 123}.Encode(),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(token.After)
	// Output: 123
}

func ExampleParseOr() {
	token, err := pagetoken.ParseOr("", pagetoken.Offset{
		ReadTimestamp: must.Ok(time.Parse(time.DateTime, "2020-01-01 10:20:30")),
		Limit:         100,
		Offset:        200,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(token.Offset)
	// Output: 200
}
