package signal

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestShutdownContext(t *testing.T) {
	ctx, cancel := GetShutdownContext(context.Background())
	wg := &sync.WaitGroup{}

	// register callbacks that are called when the cancel()-func was executed
	for i := 0; i < 10; i++ {
		if err := RegisterOnShutdownCallback(ctx, func() { wg.Done() }); err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
	}

	// wait 100 milliseconds before canceling
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// this should block
	if err := WaitForShutdownContext(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait() // this should return immediately
}
