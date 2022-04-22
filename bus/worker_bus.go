package bus

import (
	"sync"
	"time"
)

// A WorkerBus uses a queue to buffer message that are passed via Bus.Publish.
// Publishing on a WorkerBus will block, if the queue is full or otherwise return
// immediately. A WorkerBus employs go-routines to pick up queued messages and
// delivers them to Subscriber(s).
type WorkerBus[E any] interface {
	Bus[E]

	// PublishTimeout publishes a message on the Bus. If the queue is full, waits
	// a maximum amount of time before cancelling the operation. Returns true of the
	// message was enqueued, false if not.
	PublishTimeout(msg E, timeout time.Duration) bool
}

// WorkerBusSingletonQueueSize - Size of the queue used by the WorkerBus singletons
var WorkerBusSingletonQueueSize = 1000

var (
	workerBusMtx = &sync.Mutex{}
	workerBusses = map[string]WorkerBus[any]{}
)

// GetNamedWorkerBus - Provides thread-safe access to a WorkerBus singleton with
// a specified name. Repeated calls with the same name always return the same WorkerBus.
func GetNamedWorkerBus(name string) WorkerBus[any] {
	workerBusMtx.Lock()
	defer workerBusMtx.Unlock()
	if b, ok := workerBusses[name]; ok {
		return b
	}
	b := NewWorkerBus[any](WorkerBusSingletonQueueSize)
	workerBusses[name] = b
	return b
}

var (
	workerBusOnce                = &sync.Once{}
	workerBus     WorkerBus[any] = nil
)

// GetWorkerBus - Provides thread-safe access to a WorkerBus singleton.
func GetWorkerBus() WorkerBus[any] {
	workerBusOnce.Do(func() {
		workerBus = NewWorkerBus[any](WorkerBusSingletonQueueSize)
	})
	return workerBus
}

type workerBusImpl[E any] struct {
	subMtx *sync.RWMutex
	subs   []*subWithQueue[E]
	q      chan E
	qLen   int
	seq    int64
}

func NewWorkerBus[E any](queueLen int) WorkerBus[E] {
	b := &workerBusImpl[E]{
		subMtx: &sync.RWMutex{},
		subs:   []*subWithQueue[E]{},
		q:      make(chan E, queueLen),
		qLen:   queueLen,
		seq:    0,
	}
	go b.worker()
	return b
}

func (b *workerBusImpl[E]) Publish(msg E) {
	b.q <- msg
}

func (b *workerBusImpl[E]) PublishTimeout(msg E, timeout time.Duration) bool {
	t := time.NewTicker(timeout)
	defer t.Stop()
	select {
	case b.q <- msg:
		return true
	case <-t.C:
		return false
	}
}

func (b *workerBusImpl[E]) Subscribe(sub Subscriber[E]) (unsubscribe func()) {
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	b.seq++
	s := &subWithQueue[E]{id: b.seq, sub: sub, q: make(chan E, b.qLen)}
	go s.work() // start worker for this sub
	b.subs = append(b.subs, s)
	return b.unsubscribeId(b.seq)
}

func (b *workerBusImpl[E]) unsubscribeId(id int64) (unsubscribe func()) {
	return func() {
		b.subMtx.Lock()
		defer b.subMtx.Unlock()
		var subs []*subWithQueue[E]
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

func (b *workerBusImpl[E]) worker() {
	for msg := range b.q {
		b.subMtx.RLock()
		for _, sub := range b.subs {
			sub.q <- msg // multiplex message to all sub-queues
		}
		b.subMtx.RUnlock()
	}
}

type subWithQueue[E any] struct {
	id  int64
	sub Subscriber[E]
	q   chan E
}

func (s *subWithQueue[E]) work() {
	for msg := range s.q {
		s.sub(msg)
	}
}
