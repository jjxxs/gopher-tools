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

type Repeat struct {
	mtx   *sync.Mutex
	state state
	trans chan state
	t     *time.Ticker
	fn    func()
}

func NewRepeat(fn func(), duration time.Duration) (r *Repeat) {
	r = &Repeat{
		mtx:   &sync.Mutex{},
		fn:    fn,
		state: ready,
		trans: make(chan state, 1),
		t:     time.NewTicker(duration),
	}
	go r.worker()
	return
}

func (r *Repeat) Start() (this *Repeat) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.state == ready {
		r.trans <- running
	}
	return r
}

func (r *Repeat) Stop() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.state == running {
		r.trans <- ready
	}
}

func (r *Repeat) Exit() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.trans <- exited
}

func (r *Repeat) worker() {
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
