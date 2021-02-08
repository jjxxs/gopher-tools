package execution

import (
	"sync"
	"testing"
	"time"
)

func TestRepeatedTaskStart(t *testing.T) {
	wg := sync.WaitGroup{}
	count := 0
	wg.Add(10)
	rt := NewRepeatedTask(10*time.Millisecond, func() {
		count++
		if count <= 10 {
			wg.Done()
		}
	})
	tm := time.Now()
	rt.Start()
	wg.Wait()
	diff := time.Since(tm)
	// diff should be ~100ms, allow it to be in [70, 130] so this test
	// doesn't fail on systems that are under heavy load etc. (github-actions)
	if diff <= 70*time.Millisecond || diff >= 130*time.Millisecond {
		t.Fail()
	}
}

func TestRepeatedTaskStop(t *testing.T) {
	count := 0
	var rt RepeatedTask
	rt = NewRepeatedTask(50*time.Millisecond, func() {
		count++
		rt.Stop()
	})
	rt.Start()
	time.Sleep(200 * time.Millisecond)
	if count != 1 {
		t.Fail()
	}
}
