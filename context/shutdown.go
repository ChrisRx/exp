package context

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.chrisrx.dev/x/safe"
)

// ShutdownContext is a specialized context that allows adding handler
// functions when the context is marked done.
type ShutdownContext interface {
	context.Context

	// AddHandler adds a new handler function to be associated with this
	// [ShutdownContext].
	AddHandler(func())
}

// Shutdown returns a new [ShutdownContext] using [context.Background] as the
// parent context. It runs any registered handler functions when receiving any
// of the provided signals, otherwise using a default set of signals.
//
// Each received signal will cause the next handler function to be executed
// until:
//  1. All functions have been executed.
//  2. Any of the functions complete successfully.
//
// The execution order is FIFO based on calls to [ShutdownContext.AddHandler].
func Shutdown(signals ...os.Signal) ShutdownContext {
	ctx, cancel := context.WithCancel(context.Background())
	s := &shutdownCtx{
		Context: ctx,
		ch:      make(chan os.Signal, 1),
	}
	if len(signals) == 0 {
		signals = defaultShutdownSignals
	}
	signal.Notify(s.ch, signals...)

	go func() {
		defer cancel()
		defer safe.Close(s.ch)
		defer signal.Stop(s.ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ch:
				if len(s.handlers) == 0 {
					return
				}
				var fn func()
				fn, s.handlers = s.handlers[0], s.handlers[1:]
				go func() {
					defer cancel()
					fn()
				}()
			}
		}
	}()

	runtime.AddCleanup(s, func(ch chan os.Signal) {
		cancel()
		signal.Stop(ch)
	}, s.ch)
	return s
}

var defaultShutdownSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGINT,
	syscall.SIGTERM,
}

type shutdownCtx struct {
	context.Context

	ch       chan os.Signal
	handlers []func()
}

var _ context.Context = (*shutdownCtx)(nil)

// AddHandler adds a new handler function to a [ShutdownContext] to run when it
// is marked done.
func (s *shutdownCtx) AddHandler(fn func()) {
	s.handlers = append(s.handlers, fn)
}

func (s *shutdownCtx) String() string {
	return "context.ShutdownContext"
}
