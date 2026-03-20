package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"go.chrisrx.dev/x/safe"
)

// ShutdownContext is a specialized context that allows adding handler
// functions when the context is marked done.
type ShutdownContext interface {
	context.Context

	// AddHandler adds a new handler function to be associated with a
	// [ShutdownContext].
	AddHandler(func())

	// AddCleanup adds cleanup functions associated with a [ShutdownContext].
	// These are called when the context is cleaned up by the Go runtime. There
	// is no guarantee that these will run on program exit.
	AddCleanup(func())

	// Wait blocks until the context is done. This is syntactic sugar for
	// receiving from [ShutdownContext.Done].
	Wait()
}

var defaultShutdownSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGINT,
	syscall.SIGTERM,
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
// When there are no more handler functions to execute, the context is canceled
// and the default signal behavior is restored. If no handlers are given, a
// signal received will only cancel the context and restore default signal
// behavior.
func Shutdown(signals ...os.Signal) ShutdownContext {
	ctx, cancel := context.WithCancel(context.Background())
	sh := &shutdownHandlers{ch: make(chan os.Signal, 1)}
	ctx = handlers.WithValue(ctx, sh)
	s := &shutdownCtx{
		Context: ctx,
	}

	logger := logger.With(
		slog.String("type", fmt.Sprintf("%T", s)),
		slog.String("addr", fmt.Sprintf("%p", s)),
	)

	if len(signals) == 0 {
		signals = defaultShutdownSignals
	}
	signal.Notify(sh.ch, signals...)

	go func() {
		defer cancel()
		defer safe.Close(sh.ch)
		defer signal.Stop(sh.ch)
		defer runtime.GC()

		for {
			select {
			case <-ctx.Done():
				logger.Debug("parent context canceled")
				return
			case sig, ok := <-sh.ch:
				if !ok {
					logger.Debug("signal notify stopped")
					return
				}
				sh := handlers.Value(ctx)
				if len(sh.handlers) == 0 {
					logger.Debug("shutdown context done")
					// Restore normal signal handling and attempt to resend signal back
					// to this process.
					signal.Stop(sh.ch)
					self, _ := os.FindProcess(os.Getpid())
					self.Signal(sig)
					return
				}
				go func(fn func()) {
					defer cancel()
					fn()
				}(sh.next())
			}
		}
	}()

	runtime.AddCleanup(s, func(ch chan os.Signal) {
		logger.Debug("runtime cleanup called")
		cancel()
		signal.Stop(sh.ch)
		for _, fn := range handlers.Value(ctx).cleanupHandlers {
			fn()
		}
	}, sh.ch)
	return s
}

func AddHandler(ctx context.Context, fn func()) {
	switch ctx := ctx.(type) {
	case ShutdownContext:
		ctx.AddHandler(fn)
	default:
		slog.Warn("provided context is not ShutdownContext")
	}
}

func AddCleanup(ctx context.Context, fn func()) {
	switch ctx := ctx.(type) {
	case ShutdownContext:
		ctx.AddCleanup(fn)
	default:
		slog.Warn("provided context is not ShutdownContext")
	}
}

type shutdownCtx struct {
	context.Context
}

// AddHandler adds a new handler function to a [ShutdownContext] to run when it
// is marked done.
func (s *shutdownCtx) AddHandler(fn func()) {
	handlers.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addHandler(fn)
	})
}

// AddCleanup adds a new cleanup function to a [ShutdownContext] to run when it
// is garbage collected.
func (s *shutdownCtx) AddCleanup(fn func()) {
	handlers.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addCleanup(fn)
	})
}

func (s *shutdownCtx) String() string {
	return "context.ShutdownContext"
}

func (s *shutdownCtx) Wait() {
	<-s.Done()
	runtime.GC()
}

// context key for shutdown handlers
var handlers = Key[*shutdownHandlers]()

type shutdownHandlers struct {
	// for testing
	ch chan (os.Signal)

	mu              sync.Mutex
	handlers        []func()
	cleanupHandlers []func()
}

func (s *shutdownHandlers) addHandler(fn func()) {
	s.mu.Lock()
	s.handlers = append(s.handlers, fn)
	s.mu.Unlock()
}

func (s *shutdownHandlers) addCleanup(fn func()) {
	s.mu.Lock()
	s.cleanupHandlers = append(s.cleanupHandlers, fn)
	s.mu.Unlock()
}

func (s *shutdownHandlers) next() (next func()) {
	s.mu.Lock()
	next, s.handlers = s.handlers[0], s.handlers[1:]
	s.mu.Unlock()
	return
}
