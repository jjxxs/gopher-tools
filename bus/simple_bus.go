package bus

import (
	"sync"
)

// Holds 'named' Bus-singletons
var busses = &sync.Map{}

// GetNamedBus - Provides thread-safe access to a Bus with a specified
// name. Repeated calls with the same name always return the same Bus.
func GetNamedBus(name string) Bus {
	if b, ok := busses.Load(name); ok {
		return b.(Bus)
	}
	b := NewBus()
	busses.Store(name, b)
	return b
}

var (
	busOnce     = &sync.Once{}
	bus     Bus = nil
)

// GetBus - Provides thread-safe access to a Bus singleton
// which is initialized the first time GetBus is called.
// Repeated calls will return the same instance.
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

// Creates a simple Bus. No go-routines are employed by this Bus.
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
