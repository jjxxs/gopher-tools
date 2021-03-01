package bus

import (
	"reflect"
	"sync"
	"time"
)

// SingleWorkerBusMsgQueueSize - Size of the queue used by the WorkerBus
var SingleWorkerBusMsgQueueSize = 1000

// Holds 'named' WorkerBus-singletons
var (
	singleWorkerBussesMtx = &sync.Mutex{}
	singleWorkerBusses    = map[string]WorkerBus{}
)

// GetNamedSingleWorkerBus - Provides thread-safe access to a WorkerBus
// with a specified name. Repeated calls with the same name always return
// the same SingleWorkerBus.
func GetNamedSingleWorkerBus(name string) WorkerBus {
	singleWorkerBussesMtx.Lock()
	defer singleWorkerBussesMtx.Unlock()
	if b, ok := singleWorkerBusses[name]; ok {
		return b
	}
	b := NewSingleWorkerBus()
	singleWorkerBusses[name] = b
	return b
}

var (
	singleWorkerBusOnce           = &sync.Once{}
	singleWorkerBus     WorkerBus = nil
)

// GetSingleWorkerBus - Provides thread-safe access to a WorkerBus
// singleton which is initialized the first time GetSingleWorkerBus
// is called. Repeated calls will return the same instance.
func GetSingleWorkerBus() WorkerBus {
	singleWorkerBusOnce.Do(func() {
		singleWorkerBus = NewSingleWorkerBus()
	})
	return singleWorkerBus
}

type singleWorkerBusImpl struct {
	subMtx *sync.RWMutex
	subs   []Subscriber
	q      chan Message
}

// Creates a WorkerBus that uses a single go-routine to dispatch
// messages to Subscriber(s).
func NewSingleWorkerBus() WorkerBus {
	b := &singleWorkerBusImpl{
		subMtx: &sync.RWMutex{},
		subs:   []Subscriber{},
		q:      make(chan Message, SingleWorkerBusMsgQueueSize),
	}
	go b.worker()
	return b
}

func (b *singleWorkerBusImpl) Publish(msg Message) {
	b.q <- msg
}

func (b *singleWorkerBusImpl) PublishTimeout(msg Message, timeout time.Duration) bool {
	t := time.NewTicker(timeout)
	defer t.Stop()
	select {
	case b.q <- msg:
		return true
	case <-t.C:
	}
	return false
}

func (b *singleWorkerBusImpl) Subscribe(sub Subscriber) {
	s1Ptr := reflect.ValueOf(sub).Pointer()
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	for _, s2 := range b.subs {
		s2Ptr := reflect.ValueOf(s2).Pointer()
		if s1Ptr == s2Ptr {
			return // no duplicate subscribers
		}
	}
	b.subs = append(b.subs, sub)
}

func (b *singleWorkerBusImpl) Unsubscribe(sub Subscriber) {
	s1Ptr := reflect.ValueOf(sub).Pointer()
	var subs []Subscriber
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	for _, s2 := range b.subs {
		s2Ptr := reflect.ValueOf(s2).Pointer()
		if s1Ptr != s2Ptr {
			subs = append(subs, s2)
		}
	}
	b.subs = subs
}

func (b *singleWorkerBusImpl) worker() {
	for msg := range b.q {
		b.subMtx.RLock()
		for _, sub := range b.subs {
			sub.HandleMessage(msg)
		}
		b.subMtx.RUnlock()
	}
}
