package sort_test

import (
	"fmt"

	"go.chrisrx.dev/x/slices"
	"go.chrisrx.dev/x/sort"
)

func ExampleSortMap() {
	m := map[string]int{
		"A": 5,
		"B": 6,
		"C": 1,
		"D": 4,
		"E": 3,
		"F": 2,
	}
	for k, v := range slices.Reverse(sort.Map(m)).Limit(5) {
		fmt.Printf("%v: %v\n", k, v)
	}
	// Output: B: 6
	// A: 5
	// D: 4
	// E: 3
	// F: 2
}
