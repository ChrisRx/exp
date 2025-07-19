package safe

// Close safely closes a Go channel.
func Close[T any, C ~chan T](ch C) (closed bool) {
	if ch == nil {
		return
	}
	if err := Do(func() { close(ch) }); err != nil {
		if err.Error() == "close of closed channel" {
			return false
		}
		panic(err)
	}
	return true
}
