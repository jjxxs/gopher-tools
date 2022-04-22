package bus

import (
	"sync"
)

// A Subscriber is called with messages that are published on the Bus.
type Subscriber[E any] func(msg E)

// A Bus provides an implementation of a loosely-coupled publish-subscriber
// pattern. Subscriber(s) can subscribe to the Bus and are called whenever a
// Message is Publish(ed) on the Bus.
type Bus[E any] interface {
	// Publish a Message on the Bus. The Message will be forwarded to
	// all Subscriber(s).
	Publish(msg E)

	// Subscribe to the Bus. The given Subscriber will be notified
	// whenever a message is Publish(ed) on the Bus. The returned
	// function unsubscribes the given Subscriber.
	Subscribe(sub Subscriber[E]) (unsubscribe func())
}

var (
	mtx    = &sync.Mutex{}
	busses = map[string]Bus[any]{}
)

// GetNamedBus - Provides thread-safe access to a Bus singleton with a given
// name. Repeated calls with the same name always return the same Bus.
func GetNamedBus(name string) Bus[any] {
	mtx.Lock()
	defer mtx.Unlock()
	if b, ok := busses[name]; ok {
		return b
	}
	b := NewBus[any]()
	busses[name] = b
	return b
}

var (
	busOnce          = &sync.Once{}
	bus     Bus[any] = nil
)

// GetBus - Provides thread-safe access to a Bus singleton.
func GetBus() Bus[any] {
	busOnce.Do(func() {
		bus = NewBus[any]()
	})
	return bus
}

type busImpl[E any] struct {
	subMtx *sync.RWMutex
	subs   []*subWithId[E]
	seq    int64
}

// NewBus creates a simple Bus. No go-routines are employed by this Bus.
// Callers of Publish will directly invoke on all Subscribers.
func NewBus[E any]() Bus[E] {
	b := &busImpl[E]{
		subMtx: &sync.RWMutex{},
		subs:   []*subWithId[E]{},
		seq:    0,
	}
	return b
}

func (b *busImpl[E]) Publish(msg E) {
	b.subMtx.RLock()
	defer b.subMtx.RUnlock()
	for _, sub := range b.subs {
		sub.sub(msg)
	}
}

func (b *busImpl[E]) Subscribe(sub Subscriber[E]) (unsubscribe func()) {
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	b.seq++
	s := &subWithId[E]{id: b.seq, sub: sub}
	b.subs = append(b.subs, s)
	return b.unsubscribeId(b.seq)
}

func (b *busImpl[E]) unsubscribeId(id int64) (unsubscribe func()) {
	return func() {
		b.subMtx.Lock()
		defer b.subMtx.Unlock()
		var subs []*subWithId[E]
		for _, sub := range b.subs {
			if sub.id != id {
				subs = append(subs, sub)
			}
		}
		b.subs = subs
	}
}

type subWithId[E any] struct {
	id  int64
	sub Subscriber[E]
}
