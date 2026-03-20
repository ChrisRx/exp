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

func TestShutdown(t *testing.T) {
	lvl.Set(slog.LevelDebug)

	t.Run("hard shutdown", func(t *testing.T) {
		ctx := Shutdown()

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
			handlers.Value(ctx).ch <- syscall.SIGINT
			handlers.Value(ctx).ch <- syscall.SIGINT
			time.Sleep(100 * time.Millisecond)
			handlers.Value(ctx).ch <- syscall.SIGINT
		}()

		<-ctx.Done()
		assert.Eventually(t, true, &calledA, 100*time.Millisecond, "first signal")
		assert.Eventually(t, true, &calledB, 100*time.Millisecond, "second signal")
		assert.Eventually(t, true, &calledC, 100*time.Millisecond, "final signal")
	})

	t.Run("cleanup", func(t *testing.T) {
		ctx := Shutdown()

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
			handlers.Value(ctx).ch <- syscall.SIGINT
			handlers.Value(ctx).ch <- syscall.SIGINT
		}()

		ctx.Wait()
		assert.Eventually(t, true, &cleanupCalled, 100*time.Millisecond)
	})
}
