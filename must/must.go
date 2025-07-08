package must

import "log/slog"

// Ok checks the error and logs using the default logger.
func Ok[T any](v T, err error) T {
	if err != nil {
		slog.Error(err.Error())
	}
	return v
}
