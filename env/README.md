# env

```go
// Environment variables:
//   DATABASE_HOST=localhost
//   DATABASE_NAME=postgres
//   DATABASE_CONNECT_TIMEOUT=2m
//   TIMEOUT=20m
import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/jackc/pgx/v5"
    "go.chrisrx.dev/x/env"
)

type Config struct {
    Host           string        `env:"HOST"`
    Port           int           `env:"PORT" default:"5432"`
    Database       string        `env:"NAME"`
    ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" default:"30s"`
}

func (c Config) String() string {
    return fmt.Sprintf("postgresql://%s:%d/%s?sslmode=verify", c.Host, c.Port, c.Database)
}

var opts = env.MustParseAs[struct {
    Database Config
    Timeout  time.Duration `env:"TIMEOUT" default:"10m"`
}]()


func main() {
    ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
    defer cancel()
    
    conn, err := pgx.Connect(ctx, opts.Database.String())
    if err != nil {
        log.Fatal(err)
    }

    ...
}
```
