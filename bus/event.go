package bus

import (
	"reflect"
	"sync"
)

type Event string

var eventBuss = sync.Map{}

func GetNamedEventBus(name string) EventBus {
	var eventBus interface{}

	if eventBus, _ = eventBuss.Load(name); eventBus == nil {
		eventBus = NewEventBus()
		eventBuss.Store(name, eventBus)
	}

	return eventBus.(EventBus)
}

type EventBus interface {
	Publish(e Event, arg interface{})
	Subscribe(e Event, fn func(arg interface{}))
	Unsubscribe(e Event, fn func(arg interface{}))
	Close()
}

type eventBusImpl struct {
	rwMtx      *sync.RWMutex
	bufferSize int
	events     map[Event][]func(arg interface{})
}

// Creates a new EventBus
func NewEventBus() EventBus {
	b := &eventBusImpl{
		rwMtx:  &sync.RWMutex{},
		events: make(map[Event][]func(arg interface{})),
	}

	return b
}

func (b eventBusImpl) Publish(e Event, arg interface{}) {
	b.rwMtx.RLock()
	defer b.rwMtx.RUnlock()
	if cbs, ok := b.events[e]; ok {
		for _, cb := range cbs {
			cb(arg)
		}
	}
}

func (b eventBusImpl) Subscribe(e Event, fn func(arg interface{})) {
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	if _, ok := b.events[e]; !ok {
		b.events[e] = []func(arg interface{}){fn}
	} else {
		b.events[e] = append(b.events[e], fn)
	}
}

func (b eventBusImpl) Unsubscribe(e Event, fn func(arg interface{})) {
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	if events, ok := b.events[e]; ok {
		var cbs []func(interface{})
		if len(events) <= 1 {
			cbs = make([]func(interface{}), 0)
		} else {
			cbs = make([]func(interface{}), len(events)-1)
		}

		ptr1 := reflect.ValueOf(fn).Pointer()
		for _, cb := range events {
			ptr2 := reflect.ValueOf(cb).Pointer()
			if ptr1 != ptr2 {
				cbs = append(cbs, cb)
			}
		}
		b.events[e] = cbs
	}
}

func (b eventBusImpl) Close() {
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	b.events = make(map[Event][]func(arg interface{}))
}
