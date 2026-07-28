package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
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

	// Close closes the context and runs any registered cleanup functions.
	Close()
}

var defaultShutdownSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGINT,
	syscall.SIGTERM,
}

var (
	// context key for getting shutdown context
	shutdownCtxKey = Key[*shutdownCtx]()

	// context key for shutdown handlers
	shutdownHandlersKey = Key[*shutdownHandlers]()
)

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
	ctx = shutdownHandlersKey.WithValue(ctx, &shutdownHandlers{})
	s := &shutdownCtx{
		cancel: cancel,
	}
	s.Context = shutdownCtxKey.WithValue(ctx, s)

	logger := logger.With(
		slog.String("type", fmt.Sprintf("%T", s)),
		slog.String("addr", fmt.Sprintf("%p", s)),
	)

	if len(signals) == 0 {
		signals = defaultShutdownSignals
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	go func() {
		defer func() {
			logger.Debug("shutdown stopping ...")
			cancel()
			safe.Close(ch)
			signal.Stop(ch)
		}()

		for {
			select {
			case <-ctx.Done():
				logger.Debug("parent context canceled")
				return
			case sig, ok := <-ch:
				if !ok {
					logger.Debug("signal notify stopped")
					return
				}
				sh := shutdownHandlersKey.Value(ctx)
				if len(sh.handlers) == 0 {
					logger.Debug("shutdown context done")
					// Restore normal signal handling and attempt to resend signal back
					// to this process.
					signal.Stop(ch)
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
	return s
}

// AddHandler adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func AddHandler(ctx context.Context, fn func()) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.AddHandler(fn)
	})
}

// AddCleanup adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func AddCleanup(ctx context.Context, fn func()) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.AddCleanup(fn)
	})
}

type shutdownCtx struct {
	context.Context
	cancel context.CancelFunc
}

// AddHandler adds a new handler function to a [ShutdownContext] to run when it
// is marked done.
func (s *shutdownCtx) AddHandler(fn func()) {
	shutdownHandlersKey.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addHandler(fn)
	})
}

// AddCleanup adds a new cleanup function to a [ShutdownContext] to run when it
// is garbage collected or closed.
func (s *shutdownCtx) AddCleanup(fn func()) {
	shutdownHandlersKey.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addCleanup(fn)
	})
}

func (s *shutdownCtx) String() string {
	return "context.ShutdownContext"
}

func (s *shutdownCtx) Wait() {
	<-s.Done()
	shutdownHandlersKey.Value(s.Context).Close()
}

func (s *shutdownCtx) Close() {
	s.cancel()
	shutdownHandlersKey.Value(s.Context).Close()
}

type shutdownHandlers struct {
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

func (s *shutdownHandlers) Close() (next func()) {
	if s == nil {
		logger.Debug("close called on nil shutdown handlers")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, fn := range s.cleanupHandlers {
		if err := safe.Do(fn); err != nil {
			slog.Error("cleanup handler panic", slog.Any("err", err))
		}
	}
	s.cleanupHandlers = nil
	return
}
