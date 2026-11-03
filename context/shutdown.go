package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.chrisrx.dev/x/safe"
)

// ShutdownContext is a specialized context that allows adding handler
// functions when the context is marked done.
//
// It is two contexts. The ShutdownContext itself is the work context, the one
// passed to everything: the first signal cancels it, so work can wind down,
// gracefully if it likes. [ShutdownContext.Program] is the program's own
// context, its parent: it is canceled last, by [ShutdownContext.Close], after
// the cleanup functions have run. Things that must outlive the work, like a
// child process a graceful shutdown still needs, belong to the program context.
type ShutdownContext interface {
	context.Context

	// Program returns the program's context, the parent of the work context.
	// It is canceled after the work context and the cleanup functions, when
	// the [ShutdownContext] is closed.
	Program() context.Context

	// Intercept adds a function to run on an interrupt (CTRL+C) in place of
	// canceling the work context, e.g. to ask for another to confirm. Each one
	// consumes a single interrupt, in FIFO order, and only before the work
	// context is canceled. Other signals, like SIGTERM, are never intercepted.
	Intercept(fn func())

	// InterceptTimeout is like [ShutdownContext.Intercept], but the intercept
	// is restored when no other interrupt follows within timeout.
	InterceptTimeout(fn func(), timeout time.Duration)

	// AddHandler adds a handler function to run on a further signal, after the
	// first has canceled the work context, e.g. to stop at once when a graceful
	// shutdown is taking too long.
	AddHandler(func())

	// AddCleanup adds cleanup functions associated with a [ShutdownContext].
	// These are called when the context is cleaned up by the Go runtime. There
	// is no guarantee that these will run on program exit.
	AddCleanup(func())

	// Wait blocks until the context is done. This is syntactic sugar for
	// receiving from [ShutdownContext.Done].
	Wait()

	// Close cancels the work context, runs any registered cleanup functions
	// and then cancels the program context.
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
// parent of its program context. It runs any registered handler functions when
// receiving any of the provided signals, otherwise using a default set of
// signals.
//
// Interrupts are first consumed by any intercepts, see
// [ShutdownContext.Intercept]. Then the first signal cancels the work context,
// with a [SignalError] as its cause, which is where work notices it should
// stop, gracefully if it likes. Each further signal runs the next
// handler function, in FIFO order of calls to [ShutdownContext.AddHandler], to
// escalate, e.g. to a hard stop when a graceful one takes too long. Once there
// are no handlers left, a further signal restores the default signal behavior
// and is raised again, ending the process.
//
// So with no handlers, the first signal lets a program that waits on the
// context exit normally, and a second ends it.
func Shutdown(signals ...os.Signal) ShutdownContext {
	program, cancelProgram := context.WithCancel(context.Background())
	s := &shutdownCtx{
		cancelProgram: cancelProgram,
	}
	// The values are set on the program context, so they're found from either.
	program = shutdownHandlersKey.WithValue(program, &shutdownHandlers{})
	program = shutdownCtxKey.WithValue(program, s)
	s.program = program
	s.Context, s.cancel = context.WithCancelCause(program)
	ctx, cancel := s.Context, s.cancel

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
			cancel(nil)
			safe.Close(ch)
			signal.Stop(ch)
		}()

		for {
			select {
			case <-program.Done():
				logger.Debug("program context canceled")
				return
			case sig, ok := <-ch:
				if !ok {
					logger.Debug("signal notify stopped")
					return
				}
				sh := shutdownHandlersKey.Value(ctx)
				if ctx.Err() == nil {
					if sig == os.Interrupt {
						if fn := sh.intercept(); fn != nil {
							logger.Debug("interrupt intercepted")
							go run("shutdown intercept", fn)
							continue
						}
					}
					logger.Debug("canceling work context")
					cancel(SignalError{Signal: sig})
					continue
				}
				if sh.remaining() > 0 {
					go run("shutdown handler", sh.next())
					continue
				}
				logger.Debug("shutdown context done")
				// Restore normal signal handling and attempt to resend signal back
				// to this process.
				signal.Stop(ch)
				self, _ := os.FindProcess(os.Getpid())
				self.Signal(sig)
				return
			}
		}
	}()
	return s
}

func run(name string, fn func()) {
	if err := safe.Do(fn); err != nil {
		slog.Error(name+" panic", slog.Any("err", err))
	}
}

// Intercept adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func Intercept(ctx context.Context, fn func()) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.Intercept(fn)
	})
}

// InterceptTimeout adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func InterceptTimeout(ctx context.Context, fn func(), timeout time.Duration) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.InterceptTimeout(fn, timeout)
	})
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
	context.Context // the work context
	cancel          context.CancelCauseFunc

	program       context.Context
	cancelProgram context.CancelFunc
}

func (s *shutdownCtx) Program() context.Context {
	return s.program
}

// Intercept adds a function to a [ShutdownContext] to run on an interrupt in
// place of canceling the work context.
func (s *shutdownCtx) Intercept(fn func()) {
	s.InterceptTimeout(fn, 0)
}

// InterceptTimeout adds a function to a [ShutdownContext] to run on an
// interrupt in place of canceling the work context, restored when no other
// interrupt follows within timeout.
func (s *shutdownCtx) InterceptTimeout(fn func(), timeout time.Duration) {
	shutdownHandlersKey.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addIntercept(fn, timeout)
	})
}

// AddHandler adds a handler function to a [ShutdownContext] to run on a further
// signal, after the first has canceled the work context.
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
	s.cancelProgram()
}

func (s *shutdownCtx) Close() {
	s.cancel(nil)
	shutdownHandlersKey.Value(s.Context).Close()
	s.cancelProgram()
}

type shutdownHandlers struct {
	mu              sync.Mutex
	intercepts      []intercept
	armed           int    // the index of the next intercept
	interrupts      uint64 // counts interrupts, to tell if one followed an intercept
	handlers        []func()
	cleanupHandlers []func()
}

type intercept struct {
	fn      func()
	timeout time.Duration
}

func (s *shutdownHandlers) addIntercept(fn func(), timeout time.Duration) {
	s.mu.Lock()
	s.intercepts = append(s.intercepts, intercept{fn: fn, timeout: timeout})
	s.mu.Unlock()
}

// intercept consumes the next intercept for an interrupt, returning its
// function, or nil when there are none left.
func (s *shutdownHandlers) intercept() func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interrupts++
	if s.armed >= len(s.intercepts) {
		return nil
	}
	i, in := s.armed, s.intercepts[s.armed]
	s.armed++
	if in.timeout > 0 {
		interrupts := s.interrupts
		time.AfterFunc(in.timeout, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.interrupts == interrupts {
				s.armed = i
			}
		})
	}
	return in.fn
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

func (s *shutdownHandlers) remaining() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.handlers)
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
