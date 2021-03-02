package bus

import (
	"sync"
	"time"
)

// WorkerBusMsgQueueSize - Size of the queue used by the WorkerBus singletons
var WorkerBusMsgQueueSize = 1000

// Holds 'named' WorkerBus-singletons
var workerBusses = &sync.Map{}

// GetNamedWorkerBus - Provides thread-safe access to a WorkerBus with
// a specified name. Repeated calls with the same name always return the
// same WorkerBus.
func GetNamedWorkerBus(name string) WorkerBus {
	if b, ok := workerBusses.Load(name); ok {
		return b.(WorkerBus)
	}
	b := NewWorkerBus(WorkerBusMsgQueueSize)
	workerBusses.Store(name, b)
	return b
}

var (
	workerBusOnce           = &sync.Once{}
	workerBus     WorkerBus = nil
)

// Creates a WorkerBus that uses a go-routine for every subscriber.
func GetWorkerBus() WorkerBus {
	workerBusOnce.Do(func() {
		workerBus = NewWorkerBus(WorkerBusMsgQueueSize)
	})
	return workerBus
}

type workerBusImpl struct {
	subMtx *sync.RWMutex
	subs   []*subWithQueue
	q      chan Message
	qLen   int
	seq    int64
}

func NewWorkerBus(queueLen int) WorkerBus {
	b := &workerBusImpl{
		subMtx: &sync.RWMutex{},
		subs:   []*subWithQueue{},
		q:      make(chan Message, queueLen),
		qLen:   queueLen,
		seq:    0,
	}
	go b.worker()
	return b
}

func (b *workerBusImpl) Publish(msg Message) {
	b.q <- msg
}

func (b *workerBusImpl) PublishTimeout(msg Message, timeout time.Duration) bool {
	t := time.NewTicker(timeout)
	defer t.Stop()
	select {
	case b.q <- msg:
		return true
	case <-t.C:
		return false
	}
}

func (b *workerBusImpl) Subscribe(sub Subscriber) (unsubscribe func()) {
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	b.seq++
	s := &subWithQueue{id: b.seq, sub: sub, q: make(chan Message, b.qLen)}
	go s.work() // start worker for this sub
	b.subs = append(b.subs, s)
	return b.unsubscribeId(b.seq)
}

func (b *workerBusImpl) unsubscribeId(id int64) (unsubscribe func()) {
	return func() {
		b.subMtx.Lock()
		defer b.subMtx.Unlock()
		var subs []*subWithQueue
		for _, sub := range b.subs {
			if sub.id != id {
				subs = append(subs, sub)
			} else {
				close(sub.q) // close subs channel with stop its worker
			}
		}
		b.subs = subs
	}
}

func (b *workerBusImpl) worker() {
	for msg := range b.q {
		b.subMtx.RLock()
		for _, sub := range b.subs {
			sub.q <- msg // multiplex message to all sub-queues
		}
		b.subMtx.RUnlock()
	}
}

type subWithQueue struct {
	id  int64
	sub Subscriber
	q   chan Message
}

func (s *subWithQueue) work() {
	for msg := range s.q {
		s.sub(msg)
	}
}
