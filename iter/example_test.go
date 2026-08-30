package iter_test

import (
	"fmt"
	"strconv"

	"go.chrisrx.dev/x/iter"
)

func ExampleStream() {
	values := iter.From(func(yield func(int) bool) {
		for i := range 10 {
			if !yield(i) {
				return
			}
		}
	}).Filter(func(i int) bool {
		return i%2 == 0
	}).Map(func(i int) string {
		return strconv.Itoa(i)
	}).Collect()

	fmt.Printf("%#v\n", values)
	// Output: []string{"0", "2", "4", "6", "8"}
}
