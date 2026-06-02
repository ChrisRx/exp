package context

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/log/slog"
)

func sendSignal(sig syscall.Signal) {
	// Adding a small amount of wait when sending signals since rapidly re-sent
	// signals might get blocked instead of queued.
	time.Sleep(10 * time.Millisecond)
	syscall.Kill(os.Getpid(), sig)
}

func TestShutdown(t *testing.T) {
	lvl.Set(slog.LevelDebug)

	t.Run("non-default signal", func(t *testing.T) {
		ctx := Shutdown(syscall.SIGUSR1)
		defer ctx.Close()

		var called bool
		ctx.AddHandler(func() {
			called = true
		})

		go sendSignal(syscall.SIGUSR1)

		<-ctx.Done()
		assert.Eventually(t, true, &called, 100*time.Millisecond)
	})

	t.Run("multiple non-default signals", func(t *testing.T) {
		ctx := Shutdown(syscall.SIGUSR1, syscall.SIGUSR2)
		defer ctx.Close()

		var calledA, calledB bool
		ctx.AddHandler(func() {
			calledA = true
			select {} // will never finish
		})
		ctx.AddHandler(func() {
			calledB = true
		})

		go func() {
			sendSignal(syscall.SIGUSR1)
			sendSignal(syscall.SIGUSR2)
		}()

		<-ctx.Done()
		assert.Eventually(t, true, &calledA, 100*time.Millisecond, "first signal")
		assert.Eventually(t, true, &calledB, 100*time.Millisecond, "second signal")
	})

	t.Run("hard shutdown", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		var calledA bool
		ctx.AddHandler(func() {
			calledA = true
			fmt.Println("\rCTRL+C pressed, attempting graceful shutdown ...")
			select {} // will never finish
		})
		var calledB bool
		ctx.AddHandler(func() {
			calledB = true
			fmt.Println("\rCTRL+C pressed again, shutting down ...")
			time.Sleep(5 * time.Second) // takes approximately 5 seconds to exit
			os.Exit(1)
		})
		var calledC bool
		ctx.AddHandler(func() {
			calledC = true
			fmt.Println("\rExiting immediately")
		})

		go func() {
			sendSignal(syscall.SIGINT)
			sendSignal(syscall.SIGINT)
			time.Sleep(100 * time.Millisecond)
			sendSignal(syscall.SIGINT)
		}()

		<-ctx.Done()
		assert.Eventually(t, true, &calledA, 100*time.Millisecond, "first signal")
		assert.Eventually(t, true, &calledB, 100*time.Millisecond, "second signal")
		assert.Eventually(t, true, &calledC, 100*time.Millisecond, "final signal")
	})

	t.Run("cleanup", func(t *testing.T) {
		ctx := Shutdown()
		defer ctx.Close()

		ctx.AddHandler(func() {
			fmt.Println("\rCTRL+C pressed, attempting graceful shutdown ...")
			select {} // will never finish
		})
		ctx.AddHandler(func() {
			fmt.Println("\rCTRL+C pressed again, shutting down ...")
		})
		var cleanupCalled bool
		ctx.AddCleanup(func() {
			cleanupCalled = true
			fmt.Printf("cleaning up ...\n")
		})

		go func() {
			sendSignal(syscall.SIGINT)
			sendSignal(syscall.SIGINT)
		}()

		ctx.Wait()
		assert.Eventually(t, true, &cleanupCalled, 100*time.Millisecond)
	})
}
