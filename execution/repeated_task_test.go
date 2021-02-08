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
	rt.Start()
	wg.Wait()
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
