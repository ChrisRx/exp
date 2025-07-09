# exp

[![Go Reference](https://pkg.go.dev/badge/go.chrisrx.dev/x.svg)](https://pkg.go.dev/go.chrisrx.dev/x)

This repository contains experimental packages currently being evaluated. The goal is to cultivate highly reusable, useful and composable packages. For the sake of simplicity, all packages are contained in a single Go module and imported as `go.chrisrx.dev/x/<subpackage>`.

When a package proves useful and the API stabilizes, it will be considered for promotion to its own Go module. For example, packages like [ptr](ptr) are reasonably stable and have proven the test of time and will eventually be moved to another repository in a new Go module, changing the import path:

```go
// old
import "go.chrisrx.dev/x/ptr"

// new
import "go.chrisrx.dev/ptr"
```
