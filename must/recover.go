package must

import (
	"log/slog"
	"strings"

	"go.chrisrx.dev/x/stack"
)

func Recover() {
	if r := recover(); r != nil {
		slog.Error("panic",
			slog.String("loc", getLocation()),
			slog.Any("err", r),
		)
	}
}

const maxStackDepth = 10

func getLocation() string {
	for i := 1; i < maxStackDepth; i++ {
		s := stack.GetSource(i + 1)
		if ignoreSource(s) {
			continue
		}
		return s.String()
	}
	return "<unknown>"
}

func ignoreSource(s stack.Source) bool {
	return strings.HasPrefix(s.Name, "runtime") ||
		strings.HasPrefix(s.Name, "must.Close")
}
