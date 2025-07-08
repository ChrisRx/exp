package must

import (
	"fmt"
	"log/slog"
)

// Close safely closes a Go channel.
func Close[T any, C ~chan T](ch C) {
	if ch == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic",
				slog.String("loc", getLocation()),
				slog.Any("err", r),
				slog.String("addr", fmt.Sprintf("%v", ch)),
			)
		}
	}()
	close(ch)
}
