package context_test

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"go.chrisrx.dev/x/context"
)

// server stands in for a server with clients connected to it. Its process,
// like one started with exec.CommandContext, runs until its context is
// canceled.
type server struct {
	process  context.Context
	stop     context.CancelFunc
	exited   chan struct{}
	draining chan struct{}
}

func startServer(ctx context.Context) *server {
	process, stop := context.WithCancel(ctx)
	s := &server{process: process, stop: stop, exited: make(chan struct{}), draining: make(chan struct{})}
	go func() {
		<-process.Done()
		close(s.exited)
	}()
	return s
}

// Run serves until ctx is canceled, then stops gracefully: it drains the
// connections before stopping the process.
func (s *server) Run(ctx context.Context) {
	<-ctx.Done()
	fmt.Println("work canceled: draining connections")
	fmt.Printf("  process running: %v\n", s.process.Err() == nil)
	close(s.draining)
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("connections drained")
	case <-s.process.Done():
		fmt.Println("process stopped before the connections drained")
	}
	s.stop()
	<-s.exited
	fmt.Println("server stopped")
}

// HardStop stops the process at once.
func (s *server) HardStop() {
	s.stop()
	<-s.exited
}

// A server stops gracefully when the first signal cancels the work context,
// draining its connections. Its process belongs to the program context, so
// cancellation doesn't kill it before then. If draining takes too long, a
// second signal runs the handler, which stops the process at once. The program
// context is canceled last, by Close.
func ExampleShutdown_childProcess() {
	ctx := context.Shutdown(context.Signals(syscall.SIGUSR1))

	server := startServer(ctx.Program())
	stopped := make(chan struct{})
	go func() {
		server.Run(ctx)
		close(stopped)
	}()

	// The second signal: draining is taking too long, so stop now.
	ctx.AddHandler(func(os.Signal) {
		fmt.Println("second signal: stopping the server now")
		server.HardStop()
	})

	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	<-server.draining
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	<-stopped
	fmt.Printf("program canceled: %v\n", ctx.Program().Err() != nil)
	ctx.Close()
	fmt.Printf("program canceled: %v\n", ctx.Program().Err() != nil)

	// Output:
	// work canceled: draining connections
	//   process running: true
	// second signal: stopping the server now
	// process stopped before the connections drained
	// server stopped
	// program canceled: false
	// program canceled: true
}

// A command run in a terminal guards against an accidental CTRL+C. The first
// is intercepted and only asks for another, within 5 seconds, to confirm. The
// second cancels the work context, and AfterSignal says how it's stopping,
// which differs for a SIGTERM since none follows it. A third CTRL+C runs the
// handler, to stop at once.
func ExampleShutdown_confirmInterrupt() {
	ctx := context.Shutdown()
	defer ctx.Close()

	step := make(chan struct{})
	ctx.Intercept(func() {
		fmt.Println("CTRL+C received, press again to shut down gracefully ...")
		step <- struct{}{}
	}, 5*time.Second)

	context.AfterSignal(ctx, func(sig os.Signal) {
		switch sig {
		case os.Interrupt:
			fmt.Printf("%v received, shutting down gracefully, press again to stop now ...\n", sig)
		default:
			fmt.Printf("received %v, shutting down gracefully ...\n", sig)
		}
		step <- struct{}{}
	})

	ctx.AddHandler(func(os.Signal) {
		fmt.Println("CTRL+C pressed again, shutting down immediately ...")
		step <- struct{}{}
	})

	for range 3 {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		<-step
	}

	// Output:
	// CTRL+C received, press again to shut down gracefully ...
	// interrupt received, shutting down gracefully, press again to stop now ...
	// CTRL+C pressed again, shutting down immediately ...
}

// legacyServer stands in for a server that doesn't stop gracefully when its
// context is canceled, like one whose process was started with
// exec.CommandContext on that context.
type legacyServer struct{}

func (legacyServer) GracefulStop(ctx context.Context) {
	fmt.Println("warning players, then stopping the server")
}

// A server that only stops abruptly when its context is canceled is stopped
// gracefully by a BeforeCancel function instead. It runs on the signal that
// starts the shutdown, while the work context is still live, and the work
// context is canceled once it returns or the grace period is over.
func ExampleShutdown_beforeCancel() {
	ctx := context.Shutdown(
		context.Signals(syscall.SIGUSR1),
		context.GracePeriod(30*time.Second),
	)
	defer ctx.Close()

	var server legacyServer
	ctx.BeforeCancel(func(ctx context.Context, sig os.Signal) {
		fmt.Printf("%v received, shutting down gracefully ...\n", sig)
		fmt.Printf("  work canceled: %v\n", ctx.Err() != nil)
		server.GracefulStop(ctx)
	})

	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	<-ctx.Done()
	fmt.Println("work canceled")

	// Output:
	// user defined signal 1 received, shutting down gracefully ...
	//   work canceled: false
	// warning players, then stopping the server
	// work canceled
}
