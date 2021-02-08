package execution

import (
	"sync"
	"sync/atomic"
	"time"
)

// A RepeatedTask calls a given function repeatedly in a given
// interval after Start was called and until Stop is called.
type RepeatedTask interface {
	// Start the repeated execution until Stop is called.
	Start()
	// Stop the repeated execution until Start is called.
	Stop()
	// Set/change the interval. Effective after current
	// interval and execution.
	SetInterval(interval time.Duration)
}

type repeatedTaskImpl struct {
	interval time.Duration
	task     func()
	started  *int32
	timer    *time.Timer
	stop     chan bool
	working  *sync.WaitGroup
	mtx      *sync.Mutex
}

func NewRepeatedTask(interval time.Duration, task func()) RepeatedTask {
	r := &repeatedTaskImpl{
		interval: interval,
		task:     task,
		started:  new(int32),
		timer:    time.NewTimer(interval),
		stop:     make(chan bool, 1),
		working:  &sync.WaitGroup{},
		mtx:      &sync.Mutex{},
	}

	return r
}

func (r *repeatedTaskImpl) Start() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if atomic.CompareAndSwapInt32(r.started, 0, 1) {
		r.timer.Reset(r.interval)
		go r.worker()
		r.working.Add(1)
	}
}

func (r *repeatedTaskImpl) Stop() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if atomic.CompareAndSwapInt32(r.started, 1, 0) {
		r.stop <- true
		r.working.Wait()
		if !r.timer.Stop() {
			<-r.timer.C
		}
	}
}

func (r *repeatedTaskImpl) SetInterval(interval time.Duration) {
	r.interval = interval
}

func (r *repeatedTaskImpl) worker() {
	defer r.working.Done()
	for {
		select {
		case <-r.stop:
			return
		case <-r.timer.C:
			r.timer.Reset(r.interval)
			r.task()
		}
	}
}
