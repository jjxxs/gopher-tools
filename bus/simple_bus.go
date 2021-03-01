package bus

import (
	"reflect"
	"sync"
)

// Holds 'named' Bus-singletons
var (
	simpleBussesMtx = &sync.Mutex{}
	simpleBusses    = map[string]Bus{}
)

// GetNamedSimpleBus - Provides thread-safe access to a Bus with a
// specified name. Repeated calls with the same name always return
// the same Bus.
func GetNamedSimpleBus(name string) Bus {
	simpleBussesMtx.Lock()
	defer simpleBussesMtx.Unlock()
	if b, ok := simpleBusses[name]; ok {
		return b
	}
	b := NewSimpleBus()
	simpleBusses[name] = b
	return b
}

var (
	simpleBusOnce     = &sync.Once{}
	simpleBus     Bus = nil
)

// GetSimpleBus - Provides thread-safe access to a Bus
// singleton which is initialized the first time GetSimpleBus
// is called. Repeated calls will return the same instance.
func GetSimpleBus() Bus {
	simpleBusOnce.Do(func() {
		simpleBus = NewSimpleBus()
	})
	return simpleBus
}

type simpleBusImpl struct {
	subMtx *sync.RWMutex
	subs   []Subscriber
}

// Creates a WorkerBus that uses a single go-routine to dispatch
// messages to Subscriber(s).
func NewSimpleBus() Bus {
	b := &simpleBusImpl{
		subMtx: &sync.RWMutex{},
		subs:   []Subscriber{},
	}
	return b
}

func (b *simpleBusImpl) Publish(msg Message) {
	b.subMtx.RLock()
	defer b.subMtx.RUnlock()
	for _, sub := range b.subs {
		sub.HandleMessage(msg)
	}
}

func (b *simpleBusImpl) Subscribe(sub Subscriber) {
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

func (b *simpleBusImpl) Unsubscribe(sub Subscriber) {
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
