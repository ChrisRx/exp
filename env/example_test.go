package env_test

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"go.chrisrx.dev/x/context"
	"go.chrisrx.dev/x/env"
)

var opts = env.MustParseFor[struct {
	Addr           string        `env:"ADDR" default:":8080" validate:"split_addr().port > 1024"`
	Dir            http.Dir      `env:"DIR" $default:"tempdir()"`
	ReadTimeout    time.Duration `env:"TIMEOUT" default:"2m"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" $default:"1 << 20"`
}](env.RootPrefix("FILESERVER"))

func ExampleMustParseFor() {
	_ = os.Setenv("MYSERVICE_ADDR", ":8443")

	opts := env.MustParseFor[struct {
		Addr string `env:"ADDR" default:":8080" validate:"split_addr(self).port > 1024"`
	}](env.RootPrefix("MYSERVICE"))

	fmt.Printf("MYSERVICE_ADDR: %q\n", opts.Addr)

	// Output: MYSERVICE_ADDR: ":8443"
}

func ExampleMustParseFor_fileServer() {
	ctx := context.Shutdown()

	s := &http.Server{
		Addr:           opts.Addr,
		Handler:        http.FileServer(opts.Dir),
		ReadTimeout:    opts.ReadTimeout,
		WriteTimeout:   opts.WriteTimeout,
		MaxHeaderBytes: opts.MaxHeaderBytes,
		BaseContext:    func(net.Listener) context.Context { return ctx },
	}
	ctx.AddHandler(func() {
		fmt.Println("\rCTRL+C pressed, attempting graceful shutdown ...")
		if err := s.Shutdown(ctx); err != nil {
			panic(err)
		}
	})
	log.Printf("serving %s at %s ...\n", opts.Dir, opts.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func ExamplePrint() {
	env.Print(env.MustParseFor[struct {
		Addr           string        `env:"ADDR" default:":8080" validate:"split_addr(self).port > 1024"`
		Dir            http.Dir      `env:"DIR" $default:"tempdir()"`
		ReadTimeout    time.Duration `env:"TIMEOUT" default:"2m"`
		WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
		MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" $default:"1 << 20"`
	}]())

	// Output:
	// Addr{
	//   env=ADDR
	//   default=:8080
	//   value=:8080
	// }
	// Dir{
	//   env=DIR
	//   value=/tmp
	// }
	// ReadTimeout{
	//   env=TIMEOUT
	//   default=2m
	//   value=2m0s
	// }
	// WriteTimeout{
	//   env=WRITE_TIMEOUT
	//   default=30s
	//   value=30s
	// }
	// MaxHeaderBytes{
	//   env=MAX_HEADER_BYTES
	//   value=1048576
	// }
}
