package page_test

import (
	"fmt"
	"time"

	"go.chrisrx.dev/x/must"
	"go.chrisrx.dev/x/page"
)

func ExampleNewToken() {
	t := page.NewToken[page.Cursor[struct {
		ID int
	}]]()
	t.ReadTimestamp = must.Ok(time.Parse(time.DateTime, "2020-01-01 10:20:30"))
	t.After.ID = 123

	fmt.Println(t)
	// Output: Cursor[{ID:123} read=2020-01-01 10:20:30]
}

func ExampleParseToken() {
	s := page.Offset{
		ReadTimestamp: must.Ok(time.Parse(time.DateTime, "2020-01-01 10:20:30")),
		Limit:         100,
		Offset:        200,
	}.Encode()
	t, err := page.ParseToken[page.Offset](s)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
	// Output: Offset[limit=100 offset=200 read=2020-01-01 10:20:30]
}
