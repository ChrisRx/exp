package context

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestShutdown(t *testing.T) {
	ctx := Shutdown()

	ctx.AddHandler(func() {
		fmt.Println("\rCTRL+C pressed, attempting graceful shutdown ...")
		select {} // will never finish
	})
	ctx.AddHandler(func() {
		fmt.Println("\rCTRL+C pressed again, shutting down ...")
		time.Sleep(5 * time.Second) // takes approximately 5 seconds to exit
		os.Exit(1)
	})
	ctx.AddHandler(func() {
		fmt.Println("\rExiting immediately")
	})

	go func() {
		ctx.(*shutdownCtx).ch <- syscall.SIGINT
		time.Sleep(100 * time.Millisecond)
		ctx.(*shutdownCtx).ch <- syscall.SIGINT
		time.Sleep(100 * time.Millisecond)
		ctx.(*shutdownCtx).ch <- syscall.SIGINT
	}()

	<-ctx.Done()
}
