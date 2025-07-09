package sync_test

import (
	"fmt"
	"time"

	"go.chrisrx.dev/x/sync"
)

func ExampleChan() {
	var ch sync.Chan[int]
	ch.New(1) <- 10  // initialize and write to buffered channel
	v := <-ch.Load() // atomically load new channel and read
	fmt.Println(v)
	// Output: 10
}

func ExampleChan_reset() {
	var ch sync.Chan[int]
	ch.New(1) <- 10  // initialize and write to buffered channel
	ch.Reset() <- 20 // reset without reading, and send new value
	v := <-ch.Load() // atomically load new channel and read
	fmt.Println(v)
	// Output: 20
}

func ExampleLazyChan() {
	var ch sync.LazyChan[int]
	defer ch.Close()

	go func() {
		fmt.Println(<-ch.Load())
	}()

	ch.Load() <- 10
	// Output: 10
}

func ExampleSemaphore() {
	var sem sync.Semaphore
	sem.SetLimit(2)

	go func() {
		defer sem.Release()
		time.Sleep(50 * time.Millisecond)
	}()

	sem.Acquire(3)
	fmt.Println("done")
	// Output: done
}

func ExampleWaiter() {
	var ready sync.Waiter

	go func() {
		defer ready.Done()
		time.Sleep(50 * time.Millisecond)
	}()

	ready.Wait()
	fmt.Println("done")
	// Output: done
}

func ExampleWaiter_reset() {
	var ready sync.Waiter

	go func() {
		defer ready.Done()
		time.Sleep(100 * time.Millisecond)
	}()

	ready.Wait()
	fmt.Println("done")

	ready.Reset()

	go func() {
		defer ready.Done()
		time.Sleep(100 * time.Millisecond)
	}()

	ready.Wait()
	fmt.Println("done")
	// Output: done
	// done
}

func ExampleBoundedWaitGroup() {
	var wg sync.BoundedWaitGroup
	wg.SetLimit(2)

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("done")
		}()
	}
	wg.Wait()
	// Output: done
	// done
	// done
	// done
	// done
}
