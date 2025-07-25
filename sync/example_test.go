package sync_test

import (
	"fmt"
	"slices"
	"time"

	"go.chrisrx.dev/x/sync"
)

func ExampleChan() {
	var ch sync.Chan[string]
	c := ch.New(0)

	go func() {
		defer ch.Close()
		for range 5 {
			c <- "done"
		}
	}()

	for v := range c {
		fmt.Println(v)
	}
	// Output: done
	// done
	// done
	// done
	// done
}

func ExampleChan_buffered() {
	var ch sync.Chan[int]
	ch.New(1) <- 10  // initialize and write to buffered channel
	v := <-ch.Load() // atomically load new channel and read
	fmt.Println(v)
	// Output: 10
}

func ExampleChan_reset() {
	ch := sync.NewChan[int](10)
	ch.Load() <- 10
	ch.Reset() <- 20 // reset without reading, and send new value
	v := <-ch.Recv()
	fmt.Println(v)
	// Output: 20
}

func ExampleChan_send() {
	ch := sync.NewChan[int](0)

	sent := ch.TrySend(10) // will timeout
	time.Sleep(200 * time.Millisecond)
	fmt.Println(sent)

	go func() {
		defer ch.Close()

		ch.Send(20, 30)
	}()

	for v := range ch.Recv() {
		fmt.Println(v)
	}
	// Output: false
	// 20
	// 30
}

func ExampleChan_newChan() {
	ch := sync.NewChan[int](0).Load()

	go func() {
		defer close(ch)
		ch <- 10
		ch <- 20
		ch <- 30
	}()

	for v := range ch {
		fmt.Println(v)
	}
	// Output: 10
	// 20
	// 30
}

func ExampleLazyChan() {
	var ch sync.LazyChan[string]

	go func() {
		defer ch.Close()
		for range 5 {
			ch.Load() <- "done"
		}
	}()

	for v := range ch.Load() {
		fmt.Println(v)
	}
	// Output: done
	// done
	// done
	// done
	// done
}

func ExampleSemaphore() {
	var sem sync.Semaphore
	sem.SetLimit(2)

	go func() {
		defer sem.Release()
		time.Sleep(100 * time.Millisecond)
	}()

	sem.Acquire(3)
	fmt.Println("done")
	// Output: done
}

func ExampleSeqChan() {
	ch := sync.NewSeqChan[int](0)

	go func() {
		defer ch.Close()

		ch.Send(slices.Values([]int{10, 20, 30}))
	}()

	for v := range ch.Recv() {
		fmt.Println(v)
	}
	// Output: 10
	// 20
	// 30
}

func ExampleWaiter() {
	var ready sync.Waiter

	go func() {
		defer ready.Done()
		time.Sleep(100 * time.Millisecond)
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

func ExampleWaitGroup() {
	var wg sync.WaitGroup
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
