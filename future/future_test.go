package future_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/future"
)

func TestFuture(t *testing.T) {
	v := future.New(func() (string, error) {
		n := rand.IntN(300-100+1) + 100
		time.Sleep(time.Duration(n) * time.Millisecond)
		return "result", nil
	})

	value, err := v.Get()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "result", value)
}
