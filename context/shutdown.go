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

	"golang.org/x/term"

	"go.chrisrx.dev/x/safe"
)

// ShutdownContext is a specialized context that allows adding handler
// functions when the context is marked done.
//
// It is two contexts. The ShutdownContext itself is the work context, the one
// passed to everything: it is canceled once a shutdown starts, so work can
// wind down, gracefully if it likes. [ShutdownContext.Program] is the
// program's own context, its parent: it is canceled last, by
// [ShutdownContext.Close], after the cleanup functions have run. Things that
// must outlive the work, like a child process a graceful shutdown still needs,
// belong to the program context.
type ShutdownContext interface {
	context.Context

	// Program returns the program's context, the parent of the work context.
	// It is canceled after the work context and the cleanup functions, when
	// the [ShutdownContext] is closed.
	Program() context.Context

	// Intercept adds a function to run on an interrupt (CTRL+C) in place of
	// starting the shutdown, e.g. to ask for another to confirm. Each one
	// consumes a single interrupt, in FIFO order, and only before the shutdown
	// starts. When timeout is positive, the intercept is restored if no other
	// interrupt follows within it. Other signals, like SIGTERM, are never
	// intercepted.
	//
	// Intercepts only apply when stdin is a terminal, since otherwise no one is
	// there to press CTRL+C again. Elsewhere Intercept does nothing.
	Intercept(fn func(), timeout time.Duration)

	// BeforeCancel adds a function to run on the signal that starts the
	// shutdown, before the work context is canceled, e.g. to stop a service
	// that doesn't stop gracefully on its own when its context is canceled.
	// The functions run one at a time, in FIFO order, with a context that is
	// the work context bounded by the grace period, see [GracePeriod]. The work
	// context is canceled once they have all returned or the grace period is
	// over, whichever is first. They don't run when the shutdown starts
	// without a signal, like on [ShutdownContext.Close].
	BeforeCancel(fn func(ctx context.Context, sig os.Signal))

	// AddHandler adds a handler function to run on a further signal, after the
	// shutdown has started, e.g. to stop at once when a graceful shutdown is
	// taking too long. Each one consumes a single signal, in FIFO order.
	AddHandler(fn func(sig os.Signal))

	// AddCleanup adds cleanup functions associated with a [ShutdownContext].
	// These are called when the context is cleaned up by the Go runtime. There
	// is no guarantee that these will run on program exit.
	AddCleanup(func())

	// Wait blocks until the context is done. This is syntactic sugar for
	// receiving from [ShutdownContext.Done].
	Wait()

	// Close cancels the work context, runs any registered cleanup functions
	// and then cancels the program context. When a signal has started the
	// shutdown, it first waits for the [ShutdownContext.BeforeCancel]
	// functions, up to the grace period, so it must not be called from one.
	Close()
}

var defaultShutdownSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

// DefaultGracePeriod is the grace period of a [ShutdownContext] unless set
// with [GracePeriod]. It is a little under the 30s Kubernetes waits by default
// before a SIGKILL, so the program has time to exit after the work context is
// canceled.
const DefaultGracePeriod = 25 * time.Second

// ShutdownOption configures a [ShutdownContext].
type ShutdownOption func(*shutdownConfig)

type shutdownConfig struct {
	signals []os.Signal
	grace   time.Duration
}

// Signals sets the signals that a [ShutdownContext] handles, in place of the
// default os.Interrupt and SIGTERM.
func Signals(signals ...os.Signal) ShutdownOption {
	return func(c *shutdownConfig) {
		c.signals = signals
	}
}

// GracePeriod bounds how long the [ShutdownContext.BeforeCancel] functions can
// take before the work context is canceled. It should be less than the time
// whatever sends the signal waits before killing the process, like the
// terminationGracePeriodSeconds of a Kubernetes pod. A non-positive d removes
// the limit. The default is [DefaultGracePeriod].
func GracePeriod(d time.Duration) ShutdownOption {
	return func(c *shutdownConfig) {
		c.grace = d
	}
}

var (
	// context key for getting shutdown context
	shutdownCtxKey = Key[*shutdownCtx]()

	// context key for shutdown handlers
	shutdownHandlersKey = Key[*shutdownHandlers]()
)

// isTerminal reports whether stdin is a terminal, where an interrupt can come
// from someone pressing CTRL+C. It is a variable so tests can replace it.
var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Shutdown returns a new [ShutdownContext] using [context.Background] as the
// parent of its program context. It handles os.Interrupt and SIGTERM, unless
// set otherwise with [Signals].
//
// Interrupts are first consumed by any intercepts, see
// [ShutdownContext.Intercept]. Then the first signal starts the shutdown: the
// [ShutdownContext.BeforeCancel] functions run, and then the work context is
// canceled, with a [SignalError] as its cause. That is where work notices it
// should stop, gracefully if it likes. Each further signal runs the next
// handler function, in FIFO order of calls to [ShutdownContext.AddHandler], to
// escalate, e.g. to a hard stop when a graceful one takes too long. Once there
// are no handlers left, a further signal restores the default signal behavior
// and is raised again, ending the process. Where a process can't signal
// itself, like with an interrupt on Windows, it exits instead, with 128 plus
// the signal number.
//
// So with no handlers, the first signal lets a program that waits on the
// context exit normally, and a second ends it.
func Shutdown(opts ...ShutdownOption) ShutdownContext {
	cfg := shutdownConfig{
		signals: defaultShutdownSignals,
		grace:   DefaultGracePeriod,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	program, cancelProgram := context.WithCancel(context.Background())
	s := &shutdownCtx{
		cancelProgram: cancelProgram,
		grace:         cfg.grace,
		drained:       make(chan struct{}),
	}
	// The values are set on the program context, so they're found from either.
	program = shutdownHandlersKey.WithValue(program, &shutdownHandlers{})
	program = shutdownCtxKey.WithValue(program, s)
	s.program = program
	s.Context, s.cancel = context.WithCancelCause(program)

	logger := logger.With(
		slog.String("type", fmt.Sprintf("%T", s)),
		slog.String("addr", fmt.Sprintf("%p", s)),
	)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, cfg.signals...)

	go func() {
		defer func() {
			logger.Debug("shutdown stopping ...")
			// Stop before closing, so the signal package never sends on the
			// closed channel.
			signal.Stop(ch)
			safe.Close(ch)
		}()

		sh := shutdownHandlersKey.Value(s.Context)
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
				if !sh.isStarted() {
					if sig == os.Interrupt {
						if fn := sh.intercept(); fn != nil {
							logger.Debug("interrupt intercepted")
							go run("shutdown intercept", fn)
							continue
						}
					}
					if sh.startSignal() {
						logger.Debug("starting shutdown")
						go s.drain(sig)
						continue
					}
				}
				if fn := sh.next(); fn != nil {
					go run("shutdown handler", func() { fn(sig) })
					continue
				}
				logger.Debug("shutdown context done")
				// Restore normal signal handling and attempt to resend signal back
				// to this process.
				signal.Stop(ch)
				raise(sig)
				return
			}
		}
	}()
	return s
}

// raise sends sig to this process, now that it is no longer handled, to end
// it. Where that isn't possible, like with an interrupt on Windows, it exits
// with the code a shell gives a process ended by the signal.
func raise(sig os.Signal) {
	self, err := os.FindProcess(os.Getpid())
	if err == nil {
		err = self.Signal(sig)
	}
	if err != nil {
		code := 1
		if sig, ok := sig.(syscall.Signal); ok {
			code = 128 + int(sig)
		}
		os.Exit(code)
	}
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
func Intercept(ctx context.Context, fn func(), timeout time.Duration) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.Intercept(fn, timeout)
	})
}

// BeforeCancel adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func BeforeCancel(ctx context.Context, fn func(ctx context.Context, sig os.Signal)) {
	shutdownCtxKey.ValueFunc(ctx, func(ctx *shutdownCtx) {
		ctx.BeforeCancel(fn)
	})
}

// AddHandler adds fn to the [ShutdownContext] embedded in ctx. It is a
// package-level convenience for callers that hold a plain [context.Context]
// rather than a [ShutdownContext] directly. If ctx does not contain a
// [ShutdownContext], the call is a no-op.
func AddHandler(ctx context.Context, fn func(sig os.Signal)) {
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

	grace time.Duration
	// drained is closed once a signal has started the shutdown and the work
	// context is canceled, after the BeforeCancel functions.
	drained chan struct{}
}

// drain runs the BeforeCancel functions for the signal that started the
// shutdown, then cancels the work context.
func (s *shutdownCtx) drain(sig os.Signal) {
	defer close(s.drained)
	defer s.cancel(SignalError{Signal: sig})

	ctx, cancel := s.Context, context.CancelFunc(func() {})
	if s.grace > 0 {
		ctx, cancel = context.WithTimeout(s.Context, s.grace)
	}
	defer cancel()

	sh := shutdownHandlersKey.Value(s.Context)
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Functions added while these run are run too.
		for fn := sh.nextBefore(); fn != nil && ctx.Err() == nil; fn = sh.nextBefore() {
			run("before cancel", func() { fn(ctx, sig) })
		}
	}()
	select {
	case <-done:
	case <-ctx.Done():
		logger.Debug("grace period over, canceling work context")
	}
}

func (s *shutdownCtx) Program() context.Context {
	return s.program
}

// Intercept adds a function to a [ShutdownContext] to run on an interrupt in
// place of starting the shutdown, restored when timeout is positive and no
// other interrupt follows within it.
func (s *shutdownCtx) Intercept(fn func(), timeout time.Duration) {
	if !isTerminal() {
		return
	}
	shutdownHandlersKey.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addIntercept(fn, timeout)
	})
}

// BeforeCancel adds a function to a [ShutdownContext] to run on the signal that
// starts the shutdown, before the work context is canceled.
func (s *shutdownCtx) BeforeCancel(fn func(ctx context.Context, sig os.Signal)) {
	shutdownHandlersKey.ValueFunc(s.Context, func(sh *shutdownHandlers) {
		sh.addBefore(fn)
	})
}

// AddHandler adds a handler function to a [ShutdownContext] to run on a further
// signal, after the shutdown has started.
func (s *shutdownCtx) AddHandler(fn func(sig os.Signal)) {
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
	sh := shutdownHandlersKey.Value(s.Context)
	// From here on a signal is a further one, so it runs a handler.
	if sh.startClose() {
		// The BeforeCancel functions still need the work context.
		<-s.drained
	}
	s.cancel(nil)
	sh.Close()
	s.cancelProgram()
}

type shutdownHandlers struct {
	mu              sync.Mutex
	started         bool // a signal or Close has started the shutdown
	draining        bool // a signal started it, so the BeforeCancel functions run
	intercepts      []intercept
	armed           int    // the index of the next intercept
	interrupts      uint64 // counts interrupts, to tell if one followed an intercept
	before          []func(context.Context, os.Signal)
	handlers        []func(os.Signal)
	cleanupHandlers []func()
}

type intercept struct {
	fn      func()
	timeout time.Duration
}

func (s *shutdownHandlers) isStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// startSignal starts the shutdown for a signal, reporting false when it had
// already started.
func (s *shutdownHandlers) startSignal() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return false
	}
	s.started, s.draining = true, true
	return true
}

// startClose starts the shutdown for Close, reporting whether a signal had
// already started it.
func (s *shutdownHandlers) startClose() (draining bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = true
	return s.draining
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

func (s *shutdownHandlers) addBefore(fn func(context.Context, os.Signal)) {
	s.mu.Lock()
	s.before = append(s.before, fn)
	s.mu.Unlock()
}

// nextBefore consumes the next BeforeCancel function, or returns nil when
// there are none left.
func (s *shutdownHandlers) nextBefore() (next func(context.Context, os.Signal)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.before) == 0 {
		return nil
	}
	next, s.before = s.before[0], s.before[1:]
	return next
}

func (s *shutdownHandlers) addHandler(fn func(os.Signal)) {
	s.mu.Lock()
	s.handlers = append(s.handlers, fn)
	s.mu.Unlock()
}

// next consumes the next handler, or returns nil when there are none left.
func (s *shutdownHandlers) next() (next func(os.Signal)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.handlers) == 0 {
		return nil
	}
	next, s.handlers = s.handlers[0], s.handlers[1:]
	return next
}

func (s *shutdownHandlers) addCleanup(fn func()) {
	s.mu.Lock()
	s.cleanupHandlers = append(s.cleanupHandlers, fn)
	s.mu.Unlock()
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
