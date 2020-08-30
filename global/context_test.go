package global

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestShutdownContext(t *testing.T) {
	ctx, cancel := GetShutdownContext(context.Background())
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		if err := RegisterOnShutdownCallback(ctx, func() { wg.Done() }); err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
	}

	var shutdown time.Time
	go func() {
		if err := WaitForShutdownContext(ctx); err != nil {
			t.Fatal(err)
		}
		shutdown = time.Now()
	}()

	time.Sleep(1 * time.Second)
	cancel()
	wg.Wait()
	diff := time.Since(shutdown)
	t.Logf("diff was %d milliseconds %d microseconds %d nanoseconds",
		diff.Milliseconds(), diff.Microseconds(), diff.Nanoseconds())
	if diff.Milliseconds() > 100 {
		t.Fatal("diff was high, does blocking work for shutdown-context?")
	}
}
