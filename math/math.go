package math

import "golang.org/x/exp/constraints"

func Sum[N constraints.Integer](S ...N) (result N) {
	for _, n := range S {
		result += n
	}
	return
}
