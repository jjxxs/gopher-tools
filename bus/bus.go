package bus

import (
	"reflect"
	"sync"
)

// Holds named bus-singletons
var buss = sync.Map{}

// GetNamedBus - Provides thread-safe access to buses with a specified name.
// This can be used as a store of bus singletons. Calling the function
// with the same name always returns the same reference. Different names
// always return different references.
func GetNamedBus(name string) Bus {
	var bus interface{}

	if bus, _ = buss.Load(name); bus == nil {
		bus = NewBus()
		buss.Store(name, bus)
	}

	return bus.(Bus)
}

// QueueSize - Size of the buffer for published messages per subscriber
var QueueSize = 100

// A Bus provides a implementation of a loosely-coupled publish-subscriber
// pattern. Interested consumers can subscribe with a function that will
// be called with the published message.
// Every subscriber has a dedicated go-routine that will deliver the message
// to the subscriber and a queue is used to buffer pending messages.
type Bus interface {
	// Publishes a message to the bus. For every subscriber, the message
	// is put into the subscribers queue and delivered in a go-routine dedicated
	// to that specific subscriber. If a subscribers message-queue is full,
	// publish will block until it can be added.
	Publish(arg interface{})

	// Subscribe to the bus. The provided function will be called from a
	// dedicated go-routine.
	Subscribe(fn func(arg interface{}))

	// Unsubscribe from the bus.
	Unsubscribe(fn func(arg interface{}))

	// Closes the bus. All subscribers will be removed from the bus.
	Close()
}

type handler struct {
	callback func(arg interface{})
	queue    chan interface{}
}

type busImpl struct {
	rwMtx     sync.RWMutex
	queueSize int
	handlers  []*handler
}

// Creates a new Bus
func NewBus() Bus {
	b := &busImpl{
		rwMtx:     sync.RWMutex{},
		queueSize: QueueSize,
		handlers:  make([]*handler, 0),
	}

	return b
}

func (b *busImpl) Subscribe(fn func(arg interface{})) {
	// create new handler
	h := &handler{
		callback: fn,
		queue:    make(chan interface{}, b.queueSize),
	}

	// call handler with arg whenever something is published on the bus
	go func() {
		for args := range h.queue {
			h.callback(args)
		}
	}()

	// add handler
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	b.handlers = append(b.handlers, h)
}

func (b *busImpl) Publish(arg interface{}) {
	b.rwMtx.RLock()
	defer b.rwMtx.RUnlock()
	for _, handler := range b.handlers {
		handler.queue <- arg
	}
}

func (b *busImpl) Unsubscribe(fn func(arg interface{})) {
	var hs []*handler

	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()

	ptr1 := reflect.ValueOf(fn).Pointer()
	for _, h := range b.handlers {
		ptr2 := reflect.ValueOf(h.callback).Pointer()
		if ptr1 == ptr2 {
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

	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	b.handlers = make([]*handler, 0)
}
