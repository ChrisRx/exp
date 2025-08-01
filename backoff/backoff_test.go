package backoff

import (
	"fmt"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	var bo Backoff

	bo.Jitter = 10 * time.Millisecond

	for range 30 {
		fmt.Printf("bo.Next(): %v\n", bo.Next())
	}
}
