package context

import (
	"context"
	"errors"
	"os"
)

// SignalError is the cause of a context canceled by a signal, as returned by
// [Cause], e.g. for the work context of a [ShutdownContext].
type SignalError struct {
	Signal os.Signal
}

func (e SignalError) Error() string {
	return e.Signal.String() + " signal received"
}

// Is reports a SignalError as [Canceled], so errors.Is(Cause(ctx), Canceled)
// holds, like it does for the cause from [os/signal.NotifyContext].
func (e SignalError) Is(target error) bool {
	return target == context.Canceled
}

// SignalCause returns the signal that canceled ctx, if it was canceled by one.
func SignalCause(ctx context.Context) (os.Signal, bool) {
	var err SignalError
	if errors.As(context.Cause(ctx), &err) {
		return err.Signal, true
	}
	return nil, false
}

// AfterSignal is like [AfterFunc], but calls f only when ctx was canceled by a
// signal, with that signal. When ctx is canceled for another reason, like
// [ShutdownContext.Close] on a normal exit, f isn't called.
func AfterSignal(ctx context.Context, f func(os.Signal)) (stop func() bool) {
	return context.AfterFunc(ctx, func() {
		if sig, ok := SignalCause(ctx); ok {
			f(sig)
		}
	})
}
