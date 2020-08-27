package interval

import (
	"fmt"
	"testing"
	"time"
)

func TestRepeatedTaskStartStop(t *testing.T) {
	var last = time.Now()
	var i = 0

	var task = func() {
		i++
		var diff = time.Since(last)
		last = time.Now()
		fmt.Printf("diff was %s\n", diff)
	}
	var intervalTask = NewRepeatedTask(300*time.Millisecond, task)
	intervalTask.Start()

	time.Sleep(1 * time.Second)
	intervalTask.Stop()
	if i != 4 {
		t.Fatal()
	}
}
