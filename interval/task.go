package interval

import (
	"sync"
	"time"
)

type RepeatedTask interface {
	Start()
	Stop()
	SetInterval(interval time.Duration)
}

type repeatedTaskImpl struct {
	interval  time.Duration
	task      func()
	stop      chan bool
	startOnce *sync.Once
	stopOnce  *sync.Once
	wg        *sync.WaitGroup
}

func NewRepeatedTask(interval time.Duration, task func()) RepeatedTask {
	r := &repeatedTaskImpl{
		interval:  interval,
		task:      task,
		stop:      make(chan bool, 1),
		startOnce: &sync.Once{},
		stopOnce:  &sync.Once{},
		wg:        &sync.WaitGroup{},
	}

	r.wg.Add(1)

	go func() {
		r.wg.Wait()
		t := time.NewTimer(r.interval)
		defer t.Stop()
		r.task() // run task on start with no delay
		for {
			select {
			case <-r.stop:
				return
			case <-t.C:
				t.Reset(r.interval)
				r.task()
				break
			}
		}
	}()

	return r
}

func (r *repeatedTaskImpl) Start() {
	r.startOnce.Do(func() {
		r.wg.Done()
	})
}

func (r *repeatedTaskImpl) Stop() {
	r.stopOnce.Do(func() {
		r.stop <- true
	})
}

func (r *repeatedTaskImpl) SetInterval(interval time.Duration) {
	r.interval = interval
}
