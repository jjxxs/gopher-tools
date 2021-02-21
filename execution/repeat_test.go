package execution

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestRepeat(t *testing.T) {
	wg := sync.WaitGroup{}
	count := 0
	wg.Add(10)
	rt := NewRepeat(func() {
		count++
		if count <= 10 {
			wg.Done()
		}
	}, 10*time.Millisecond)
	tm := time.Now()
	rt.Start()
	wg.Wait()
	diff := time.Since(tm)

	// ignore the following timing requirements completely for github-actions
	if os.Getenv("GITHUB_WORKFLOW") != "" {
		return
	}

	// diff should be ~100ms, allow it to be in [90, 110] so this test
	// doesn't fail on systems that are under heavy load etc.
	if diff <= 90*time.Millisecond || diff >= 110*time.Millisecond {
		t.Fail()
	}
}

func TestRepeatedTaskStop(t *testing.T) {
	count := 0
	var rt *repeat
	rt = NewRepeat(func() {
		count++
		rt.Stop()
	}, 50*time.Millisecond)
	rt.Start()
	time.Sleep(200 * time.Millisecond)
	if count != 1 {
		t.Fail()
	}
}
