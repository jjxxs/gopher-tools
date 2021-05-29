package bus

import "sync"

// Message type of the Message carried by this Bus - change it to a specific type if necessary.
type Message = interface{}

// A Subscriber is called with messages that are published on the Bus.
type Subscriber func(msg Message)

// A Bus provides an implementation of a loosely-coupled publish-subscriber
// pattern. Subscriber(s) can subscribe to the Bus and are called whenever a
// Message is Publish(ed) on the Bus.
type Bus interface {
	// Publish a Message on the Bus. The Message will be forwarded to
	// all Subscriber(s).
	Publish(msg Message)

	// Subscribe to the Bus. The given Subscriber will be notified
	// whenever a message is Publish(ed) on the Bus. The returned
	// function unsubscribes the given Subscriber.
	Subscribe(sub Subscriber) (unsubscribe func())
}

var (
	mtx    = &sync.Mutex{}
	busses = map[string]Bus{}
)

// GetNamedBus - Provides thread-safe access to a Bus singleton with a given
// name. Repeated calls with the same name always return the same Bus.
func GetNamedBus(name string) Bus {
	mtx.Lock()
	defer mtx.Unlock()
	if b, ok := busses[name]; ok {
		return b
	}
	b := NewBus()
	busses[name] = b
	return b
}

var (
	busOnce     = &sync.Once{}
	bus     Bus = nil
)

// GetBus - Provides thread-safe access to a Bus singleton.
func GetBus() Bus {
	busOnce.Do(func() {
		bus = NewBus()
	})
	return bus
}

type busImpl struct {
	subMtx *sync.RWMutex
	subs   []*subWithId
	seq    int64
}

// NewBus creates a simple Bus. No go-routines are employed by this Bus.
// Callers of Publish will directly invoke on all Subscribers.
func NewBus() Bus {
	b := &busImpl{
		subMtx: &sync.RWMutex{},
		subs:   []*subWithId{},
		seq:    0,
	}
	return b
}

func (b *busImpl) Publish(msg Message) {
	b.subMtx.RLock()
	defer b.subMtx.RUnlock()
	for _, sub := range b.subs {
		sub.sub(msg)
	}
}

func (b *busImpl) Subscribe(sub Subscriber) (unsubscribe func()) {
	b.subMtx.Lock()
	defer b.subMtx.Unlock()
	b.seq++
	s := &subWithId{id: b.seq, sub: sub}
	b.subs = append(b.subs, s)
	return b.unsubscribeId(b.seq)
}

func (b *busImpl) unsubscribeId(id int64) (unsubscribe func()) {
	return func() {
		b.subMtx.Lock()
		defer b.subMtx.Unlock()
		var subs []*subWithId
		for _, sub := range b.subs {
			if sub.id != id {
				subs = append(subs, sub)
			}
		}
		b.subs = subs
	}
}

type subWithId struct {
	id  int64
	sub Subscriber
}
