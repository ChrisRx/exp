package context

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/log/slog"
)

func init() {
	// Tests don't run in a terminal, but intercepts are tested as if they did.
	isTerminal = func() bool { return true }
}

func sendSignal(sig syscall.Signal) {
	// Adding a small amount of wait when sending signals since rapidly re-sent
	// signals might get blocked instead of queued.
	time.Sleep(10 * time.Millisecond)
	syscall.Kill(os.Getpid(), sig)
}

// receive waits for a value on ch and returns it, failing the test after
// timeout.
func receive[T any](t *testing.T, ch <-chan T, timeout time.Duration, msg string) (v T) {
	t.Helper()
	select {
	case v = <-ch:
	case <-time.After(timeout):
		t.Fatal(msg)
	}
	return v
}

func TestShutdown(t *testing.T) {
	lvl.Set(slog.LevelDebug)

	t.Run("non-default signal", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		called := make(chan struct{}, 1)
		ctx.AddHandler(func(os.Signal) {
			called <- struct{}{}
		})

		// The first signal cancels the work context, and the second runs the handler.
		go sendSignal(syscall.SIGUSR1)
		<-ctx.Done()
		assert.Equal(t, 0, len(called), "handler run on the first signal")
		go sendSignal(syscall.SIGUSR1)
		receive(t, called, 100*time.Millisecond, "the second signal didn't run the handler")
	})

	t.Run("multiple non-default signals", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1, syscall.SIGUSR2))
		defer ctx.Close()

		calledA, calledB := make(chan struct{}, 1), make(chan struct{}, 1)
		ctx.AddHandler(func(os.Signal) {
			calledA <- struct{}{}
			select {} // will never finish
		})
		ctx.AddHandler(func(os.Signal) {
			calledB <- struct{}{}
		})

		go func() {
			sendSignal(syscall.SIGUSR1)
			sendSignal(syscall.SIGUSR2)
			sendSignal(syscall.SIGUSR1)
		}()

		<-ctx.Done()
		receive(t, calledA, 100*time.Millisecond, "the second signal didn't run the first handler")
		receive(t, calledB, 100*time.Millisecond, "the third signal didn't run the second handler")
	})

	t.Run("hard shutdown", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		calledA, calledB := make(chan struct{}, 1), make(chan struct{}, 1)
		ctx.AddHandler(func(os.Signal) {
			calledA <- struct{}{}
			fmt.Println("\rCTRL+C pressed again, attempting a faster shutdown ...")
			select {} // will never finish
		})
		ctx.AddHandler(func(os.Signal) {
			calledB <- struct{}{}
			fmt.Println("\rCTRL+C pressed a third time, exiting immediately")
		})

		go func() {
			sendSignal(syscall.SIGINT)
			sendSignal(syscall.SIGINT)
			time.Sleep(100 * time.Millisecond)
			sendSignal(syscall.SIGINT)
		}()

		<-ctx.Done()
		receive(t, calledA, 200*time.Millisecond, "the second signal didn't run the first handler")
		receive(t, calledB, 200*time.Millisecond, "the third signal didn't run the second handler")
	})

	t.Run("cleanup", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		ctx.AddHandler(func(os.Signal) {
			fmt.Println("\rCTRL+C pressed again, shutting down ...")
		})
		var cleanupCalled bool
		ctx.AddCleanup(func() {
			cleanupCalled = true
			fmt.Printf("cleaning up ...\n")
		})

		go func() {
			sendSignal(syscall.SIGINT)
		}()

		ctx.Wait()
		assert.Eventually(t, true, &cleanupCalled, 100*time.Millisecond)
	})
}

func TestShutdownProgram(t *testing.T) {
	t.Run("escalates to a hard stop on a second signal", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		program := ctx.Program()
		hardStopped := make(chan bool, 1)
		ctx.AddHandler(func(os.Signal) {
			hardStopped <- program.Err() != nil
		})

		go sendSignal(syscall.SIGUSR1)
		<-ctx.Done()
		assert.Equal(t, nil, program.Err(), "program context canceled with the work context")

		go sendSignal(syscall.SIGUSR1)
		canceled := receive(t, hardStopped, time.Second, "the second signal didn't run the handler")
		assert.Equal(t, false, canceled, "program context canceled before the hard stop")
		ctx.Close()
		assert.Equal(t, true, program.Err() != nil, "program context canceled by Close")
	})

	t.Run("values are found from both contexts", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		var cleaned bool
		// The package-level helpers find the ShutdownContext from either.
		AddCleanup(WithValue(ctx, "derived", true), func() { cleaned = true })
		assert.Equal(t, true, shutdownCtxKey.Has(ctx.Program()))
		ctx.Close()
		assert.Equal(t, true, cleaned)
	})
}

// shutdownHelper is run as a child process by TestShutdownWithoutHandlers.
func shutdownHelper() {
	ctx := Shutdown(Signals(syscall.SIGTERM))
	fmt.Println("ready")
	switch os.Getenv("SHUTDOWN_HELPER") {
	case "wait":
		// Like a command that only waits on its context.
		<-ctx.Done()
		ctx.Close()
		os.Exit(0)
	case "ignore":
		// Like a command that never looks at its context.
		select {}
	}
}

func TestShutdownWithoutHandlers(t *testing.T) {
	if os.Getenv("SHUTDOWN_HELPER") != "" {
		shutdownHelper()
		return
	}
	run := func(t *testing.T, mode string, signals int) error {
		cmd := exec.Command(os.Args[0], "-test.run=^TestShutdownWithoutHandlers$")
		cmd.Env = append(os.Environ(), "SHUTDOWN_HELPER="+mode)
		out, err := cmd.StdoutPipe()
		assert.NoError(t, err)
		assert.NoError(t, cmd.Start())
		for s := bufio.NewScanner(out); s.Scan(); {
			if s.Text() == "ready" {
				break
			}
		}
		for range signals {
			time.Sleep(50 * time.Millisecond)
			assert.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatal("the process didn't exit")
			return nil
		}
	}

	t.Run("one signal ends a command waiting on its context", func(t *testing.T) {
		assert.NoError(t, run(t, "wait", 1))
	})

	t.Run("a second signal ends a command ignoring its context", func(t *testing.T) {
		err := run(t, "ignore", 2)
		var exit *exec.ExitError
		assert.Equal(t, true, errors.As(err, &exit), "expected the process to be ended by the signal")
		status := exit.Sys().(syscall.WaitStatus)
		assert.Equal(t, syscall.SIGTERM, status.Signal())
	})
}

func TestShutdownIntercept(t *testing.T) {
	t.Run("an interrupt is intercepted before canceling", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		intercepted := make(chan struct{}, 2)
		ctx.Intercept(func() { intercepted <- struct{}{} }, 0)

		go sendSignal(syscall.SIGINT)
		receive(t, intercepted, time.Second, "the interrupt wasn't intercepted")
		assert.Equal(t, nil, ctx.Err(), "work context canceled by an intercepted interrupt")

		go sendSignal(syscall.SIGINT)
		<-ctx.Done()
		sig, ok := SignalCause(ctx)
		assert.Equal(t, true, ok)
		assert.Equal(t, os.Interrupt, sig)
		assert.Equal(t, 0, len(intercepted), "intercepted again")
	})

	t.Run("each intercept consumes one interrupt", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		order := make(chan string, 2)
		ctx.Intercept(func() { order <- "first" }, 0)
		ctx.Intercept(func() { order <- "second" }, 0)

		go sendSignal(syscall.SIGINT)
		assert.Equal(t, "first", receive(t, order, time.Second, "the first interrupt wasn't intercepted"))
		go sendSignal(syscall.SIGINT)
		assert.Equal(t, "second", receive(t, order, time.Second, "the second interrupt wasn't intercepted"))
		assert.Equal(t, nil, ctx.Err())

		go sendSignal(syscall.SIGINT)
		<-ctx.Done()
	})

	t.Run("a timeout restores the intercept", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		intercepted := make(chan struct{}, 2)
		ctx.Intercept(func() { intercepted <- struct{}{} }, 50*time.Millisecond)

		go sendSignal(syscall.SIGINT)
		receive(t, intercepted, time.Second, "the interrupt wasn't intercepted")
		time.Sleep(100 * time.Millisecond)

		// The window lapsed, so this is intercepted again.
		go sendSignal(syscall.SIGINT)
		receive(t, intercepted, time.Second, "the intercept wasn't restored")
		assert.Equal(t, nil, ctx.Err())

		// Within the window, so this one cancels.
		go sendSignal(syscall.SIGINT)
		<-ctx.Done()
	})

	t.Run("SIGTERM is never intercepted", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		var intercepted bool
		ctx.Intercept(func() { intercepted = true }, 0)

		go sendSignal(syscall.SIGTERM)
		<-ctx.Done()
		sig, _ := SignalCause(ctx)
		assert.Equal[os.Signal](t, syscall.SIGTERM, sig)
		assert.Equal(t, false, intercepted)
	})

	t.Run("intercepts don't apply after canceling", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		intercepted := make(chan struct{}, 2)
		ctx.Intercept(func() { intercepted <- struct{}{} }, 20*time.Millisecond)
		handled := make(chan struct{}, 1)
		ctx.AddHandler(func(os.Signal) { handled <- struct{}{} })

		go sendSignal(syscall.SIGINT)
		receive(t, intercepted, time.Second, "the interrupt wasn't intercepted")
		go sendSignal(syscall.SIGINT)
		<-ctx.Done()

		// The intercept's window lapses after canceling, but it goes to the handler.
		time.Sleep(50 * time.Millisecond)
		go sendSignal(syscall.SIGINT)
		receive(t, handled, time.Second, "the interrupt after canceling didn't run the handler")
		assert.Equal(t, 0, len(intercepted), "intercepted after canceling")
	})
}

func TestAfterSignal(t *testing.T) {
	t.Run("called with the signal that canceled", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		got := make(chan os.Signal, 1)
		AfterSignal(ctx, func(sig os.Signal) { got <- sig })

		go sendSignal(syscall.SIGUSR1)
		sig := receive(t, got, time.Second, "AfterSignal wasn't called")
		assert.Equal[os.Signal](t, syscall.SIGUSR1, sig)
		assert.Equal(t, true, errors.Is(Cause(ctx), Canceled))
	})

	t.Run("not called when closed", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))

		called := make(chan struct{}, 1)
		AfterSignal(ctx, func(os.Signal) { called <- struct{}{} })
		ctx.Close()

		select {
		case <-called:
			t.Fatal("AfterSignal called without a signal")
		case <-time.After(50 * time.Millisecond):
		}
		_, ok := SignalCause(ctx)
		assert.Equal(t, false, ok)
	})
}

func TestShutdownBeforeCancel(t *testing.T) {
	t.Run("runs in order before canceling", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		order := make(chan string, 2)
		ctx.BeforeCancel(func(hook Context, sig os.Signal) {
			assert.Equal[os.Signal](t, syscall.SIGUSR1, sig)
			assert.Equal(t, nil, ctx.Err(), "work context canceled before the hook")
			_, ok := hook.Deadline()
			assert.Equal(t, true, ok, "hook context has no grace deadline")
			time.Sleep(20 * time.Millisecond)
			order <- "first"
		})
		ctx.BeforeCancel(func(Context, os.Signal) {
			order <- "second"
		})

		go sendSignal(syscall.SIGUSR1)
		<-ctx.Done()
		assert.Equal(t, "first", receive(t, order, time.Second, "the first hook didn't run"))
		assert.Equal(t, "second", receive(t, order, time.Second, "the second hook didn't run"))
		sig, ok := SignalCause(ctx)
		assert.Equal(t, true, ok)
		assert.Equal[os.Signal](t, syscall.SIGUSR1, sig)
	})

	t.Run("the grace period cancels a hook that takes too long", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1), GracePeriod(50*time.Millisecond))
		defer ctx.Close()

		hookDone := make(chan error, 1)
		ctx.BeforeCancel(func(hook Context, _ os.Signal) {
			<-hook.Done()
			hookDone <- hook.Err()
		})
		var skipped bool
		ctx.BeforeCancel(func(Context, os.Signal) { skipped = true })

		go sendSignal(syscall.SIGUSR1)
		receive(t, ctx.Done(), time.Second, "the grace period didn't cancel the work context")
		err := receive(t, hookDone, time.Second, "the hook's context wasn't canceled")
		assert.Equal(t, DeadlineExceeded, err)
		assert.Equal(t, false, skipped, "hook run after the grace period")
	})

	t.Run("a further signal runs a handler while hooks run", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))
		defer ctx.Close()

		release := make(chan struct{})
		ctx.BeforeCancel(func(Context, os.Signal) { <-release })
		handled := make(chan os.Signal, 1)
		ctx.AddHandler(func(sig os.Signal) { handled <- sig })

		go func() {
			sendSignal(syscall.SIGUSR1)
			sendSignal(syscall.SIGUSR1)
		}()
		sig := receive(t, handled, time.Second, "the second signal didn't run the handler")
		assert.Equal[os.Signal](t, syscall.SIGUSR1, sig)
		assert.Equal(t, nil, ctx.Err(), "work context canceled while the hook runs")
		close(release)
		<-ctx.Done()
	})

	t.Run("Close waits for the hooks", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))

		started := make(chan struct{})
		var finished bool
		ctx.BeforeCancel(func(Context, os.Signal) {
			close(started)
			time.Sleep(50 * time.Millisecond)
			finished = true
		})

		go sendSignal(syscall.SIGUSR1)
		receive(t, started, time.Second, "the hook didn't run")
		ctx.Close()
		assert.Equal(t, true, finished, "Close returned before the hook")
	})

	t.Run("not run on Close", func(t *testing.T) {
		ctx := Shutdown(Signals(syscall.SIGUSR1))

		var called bool
		ctx.BeforeCancel(func(Context, os.Signal) { called = true })
		ctx.Close()
		assert.Equal(t, false, called)
	})
}

func TestShutdownInterceptTerminal(t *testing.T) {
	isTerminal = func() bool { return false }
	defer func() { isTerminal = func() bool { return true } }()

	ctx := Shutdown()
	defer ctx.Close()

	var intercepted bool
	ctx.Intercept(func() { intercepted = true }, 0)

	go sendSignal(syscall.SIGINT)
	receive(t, ctx.Done(), time.Second, "the interrupt didn't cancel the work context")
	assert.Equal(t, false, intercepted, "intercepted outside a terminal")
}
