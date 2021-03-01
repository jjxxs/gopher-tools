package bus

import (
	"reflect"
	"sync"
	"time"
)

// MultiWorkerBusMsgQueueSize - Size of the queue used by the WorkerBus
var MultiWorkerBusMsgQueueSize = 1000

// Holds 'named' WorkerBus-singletons
var (
	multiWorkerBussesMtx = &sync.Mutex{}
	multiWorkerBusses    = map[string]WorkerBus{}
)

// GetNamedMultiWorkerBus - Provides thread-safe access to a MultiWorkerBus
// with a specified name. Repeated calls with the same name always return
// the same MultiWorkerBus.
func GetNamedMultiWorkerBus(name string) WorkerBus {
	multiWorkerBussesMtx.Lock()
	defer multiWorkerBussesMtx.Unlock()
	if b, ok := multiWorkerBusses[name]; ok {
		return b
	}
	b := NewMultiWorkerBus()
	multiWorkerBusses[name] = b
	return b
}

var (
	multiWorkerBusOnce           = &sync.Once{}
	multiWorkerBus     WorkerBus = nil
)

// Creates a WorkerBus that uses a go-routine for every subscriber.
func GetMultiWorkerBus() WorkerBus {
	multiWorkerBusOnce.Do(func() {
		multiWorkerBus = NewMultiWorkerBus()
	})
	return multiWorkerBus
}

type multiWorkerBusImpl struct {
	subMtx *sync.RWMutex
	subs   []subWithWorker
	msgs   chan Message
}

func (b *multiWorkerBusImpl) Stop() {
	panic("implement me")
}

func NewMultiWorkerBus() WorkerBus {
	b := &multiWorkerBusImpl{
		subMtx: &sync.RWMutex{},
		subs:   []subWithWorker{},
		msgs:   make(chan Message, MultiWorkerBusMsgQueueSize),
	}
	go b.worker()
	return b
}

func (b *multiWorkerBusImpl) Publish(msg Message) {
	b.msgs <- msg
}

func (b *multiWorkerBusImpl) PublishTimeout(msg Message, timeout time.Duration) bool {
	t := time.NewTicker(timeout)
	defer t.Stop()
	select {
	case b.msgs <- msg:
		return true
	case <-t.C:
	}
	return false
}

func (b *multiWorkerBusImpl) Subscribe(sub Subscriber) {
	var s1Id = sub
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	for _, s2 := range b.subs {
		var s2Id = s2.sub
		if &s1Id == &s2Id {
			return // no duplicate subscribers
		}
	}
	s := subWithWorker{sub: sub, q: make(chan Message, MultiWorkerBusMsgQueueSize)}
	go s.work() // every sub has its own worker
	b.subs = append(b.subs, s)
}

func (b *multiWorkerBusImpl) Unsubscribe(sub Subscriber) {
	s1Ptr := reflect.ValueOf(sub).Pointer()
	var subs []subWithWorker
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	for _, s2 := range b.subs {
		s2Ptr := reflect.ValueOf(s2.sub).Pointer()
		if s1Ptr != s2Ptr {
			subs = append(subs, s2)
		} else {
			close(s2.q) // closing channel will stop the worker
		}
	}
	b.subs = subs
}

func (b *multiWorkerBusImpl) worker() {
	for msg := range b.msgs {
		b.subMtx.RLock()
		for _, sub := range b.subs {
			sub.q <- msg // multiplex message to all sub-queues
		}
		b.subMtx.RUnlock()
	}
}

type subWithWorker struct {
	sub Subscriber
	q   chan Message
}

func (s *subWithWorker) onMessage(msg Message) {
	s.q <- msg
}

func (s *subWithWorker) work() {
	for msg := range s.q {
		s.sub.HandleMessage(msg)
	}
}
