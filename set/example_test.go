package set_test

import (
	"fmt"

	"go.chrisrx.dev/x/set"
	"go.chrisrx.dev/x/slices"
)

func ExampleSet() {
	var s set.Set[int]

	for range 10 {
		s.Add(1)
	}
	s.Add(2)

	fmt.Println(slices.Sort(s.List()))
	// Output: [1 2]
}

func ExampleSet_new() {
	s := set.New(1, 2, 2, 3, 3, 3, 4, 5, 5)

	fmt.Println(slices.Sort(s.List()))
	// Output: [1 2 3 4 5]
}

func ExampleSet_uncomparable() {
	var s set.Set[[]byte]

	s.Add([]byte("1"))
	s.Add([]byte("1"))
	s.Add([]byte("2"))
	s.Add([]byte("2"))
	s.Add([]byte("3"))

	fmt.Println(slices.Sort(slices.Map(s.List(), func(elem []byte) string {
		return string(elem)
	})))
	// Output: [1 2 3]
}
