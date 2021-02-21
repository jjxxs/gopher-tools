package execution

import (
	"sync"
	"time"
)

type state int

const (
	ready state = iota
	running
	exited
)

type repeat struct {
	mtx   *sync.Mutex
	state state
	trans chan state
	t     *time.Ticker
	fn    func()
}

func NewRepeat(fn func(), duration time.Duration) (r *repeat) {
	r = &repeat{
		mtx:   &sync.Mutex{},
		fn:    fn,
		state: ready,
		trans: make(chan state, 1),
		t:     time.NewTicker(duration),
	}
	go r.worker()
	return
}

func (r *repeat) Start() (this *repeat) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.state == ready {
		r.trans <- running
	}
	return r
}

func (r *repeat) Stop() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.state == running {
		r.trans <- ready
	}
}

func (r *repeat) Exit() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.trans <- exited
}

func (r *repeat) worker() {
	for r.state != exited {
		select {
		case s := <-r.trans:
			r.state = s
		case <-r.t.C:
			if r.state == running {
				r.fn()
			}
		}
	}
	r.t.Stop()
}
