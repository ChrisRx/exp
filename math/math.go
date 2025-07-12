package math

import "go.chrisrx.dev/x/constraints"

func Sum[N constraints.Integer](S ...N) (result N) {
	for _, n := range S {
		result += n
	}
	return
}
