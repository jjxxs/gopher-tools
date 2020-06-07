package bus

import (
	"reflect"
	"sync"
)

var (
	buss = sync.Map{}
)

// Provides thread-safe access to buses with a specified name.
// This can be used as a store of bus singletons.
func GetBusFromFactory(name string) Bus {
	var bus interface{}

	if bus, _ = buss.Load(name); bus == nil {
		bus = NewBus()
		buss.Store(name, bus)
	}

	return bus.(Bus)
}

// A Bus provides a implementation of a loosely-coupled publish-subscriber
// pattern. Interested consumers can subscribe with a function that is
// called with args whenever a producer publishes on the bus.
const QueueSize = 100

type Bus interface {
	Publish(args ...interface{})
	Subscribe(fn interface{})
	Unsubscribe(fn interface{})
	Close()
}

type handler struct {
	callback reflect.Value
	queue    chan []reflect.Value
}

type busImpl struct {
	rwMutex   sync.RWMutex
	queueSize int
	handlers  []*handler
}

func NewBus() Bus {
	b := &busImpl{
		rwMutex:   sync.RWMutex{},
		queueSize: QueueSize,
		handlers:  make([]*handler, 0),
	}

	return b
}

func (b *busImpl) Subscribe(fn interface{}) {
	// fn not a function, do nothing
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return
	}

	// create new handler
	h := &handler{
		callback: reflect.ValueOf(fn),
		queue:    make(chan []reflect.Value, b.queueSize),
	}

	// call handler with args published on the bus
	go func() {
		for args := range h.queue {
			h.callback.Call(args)
		}
	}()

	// add handler
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	b.handlers = append(b.handlers, h)
}

func (b *busImpl) Publish(args ...interface{}) {
	rArgs := buildHandlerArgs(args)

	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()

	// write args to all handlers
	for _, h := range b.handlers {
		h.queue <- rArgs
	}
}

func (b *busImpl) Unsubscribe(fn interface{}) {
	rv := reflect.ValueOf(fn)

	// fn is not a function
	if rv.Type().Kind() != reflect.Func {
		return
	}

	var hs []*handler

	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()

	for _, h := range b.handlers {
		if h.callback.Pointer() == rv.Pointer() {
			close(h.queue)
		} else {
			hs = append(hs, h)
		}
	}

	b.handlers = hs
}

func (b *busImpl) Close() {
	// close all handlers
	for _, h := range b.handlers {
		close(h.queue)
	}

	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	b.handlers = make([]*handler, 0)
}

func buildHandlerArgs(args []interface{}) []reflect.Value {
	reflectedArgs := make([]reflect.Value, len(args))

	for i := 0; i < len(args); i++ {
		reflectedArgs[i] = reflect.ValueOf(args[i])
	}

	return reflectedArgs
}
