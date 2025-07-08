package sync_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.chrisrx.dev/x/sync"
)

func TestChan(t *testing.T) {
	type S struct {
		ch sync.Chan[int]
	}

	var s S
	s.ch.New(1)
	s.ch.Load() <- 10
	assert.Equal(t, 10, <-s.ch.Load())
}
