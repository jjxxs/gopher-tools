package bus

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Type of the Message carried by this Bus - change it to a specific type if you want.
type Message = interface{}

// Type of the Receiver that subscribe to this Bus - change it to a specific type if you want.
type Receiver = func(msg Message)

// MsgQueueSize - Size of the buffer for published messages per subscriber
var MsgQueueSize = 1000

// A Bus provides an implementation of a loosely-coupled publish-subscriber
// pattern. Receiver(s) can subscribe to the Bus and are called whenever a
// Message is Publish(ed) on the Bus.
type Bus interface {
	// Publish a Message on the Bus. This will pass the Message
	// to all Receiver(s) currently subscribed to this Bus.
	Publish(msg Message)

	// Subscribe to the Bus. The given Receiver will be called
	// whenever a message is Publish(ed) on the Bus.
	Subscribe(rcv Receiver)

	// Unsubscribe from the Bus. No further Message(s) will be
	// received.
	Unsubscribe(rcv Receiver)

	// Close the Bus. This effectively Unsubscribe(s) all Receiver(s)
	// and no further Message(s) can be Publish(ed).
	Close()
}

type busImpl struct {
	rwMtx  *sync.RWMutex
	rcvs   []Receiver
	msgs   chan Message
	closed *int32
}

func NewBus() Bus {
	bus := &busImpl{
		rwMtx:  &sync.RWMutex{},
		rcvs:   []Receiver{},
		msgs:   make(chan Message, MsgQueueSize),
		closed: new(int32),
	}

	go bus.worker()

	return bus
}

func (b *busImpl) worker() {
	for msg := range b.msgs {
		b.rwMtx.RLock()
		for _, rcv := range b.rcvs {
			rcv(msg)
		}
		b.rwMtx.RUnlock()
	}
}

func (b *busImpl) Publish(msg Message) {
	if atomic.CompareAndSwapInt32(b.closed, 0, 0) {
		b.msgs <- msg
	}
}

func (b *busImpl) Subscribe(rcv Receiver) {
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	b.rcvs = append(b.rcvs, rcv)
}

func (b *busImpl) Unsubscribe(rcv Receiver) {
	var rcvs []Receiver
	rcvPtr1 := reflect.ValueOf(rcv).Pointer()

	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	for _, rcv2 := range b.rcvs {
		rcvPtr2 := reflect.ValueOf(rcv2).Pointer()
		if rcvPtr1 != rcvPtr2 {
			rcvs = append(rcvs, rcv2)
		}
	}
	b.rcvs = rcvs
}

func (b *busImpl) Close() {
	if atomic.CompareAndSwapInt32(b.closed, 0, 1) {
		b.rwMtx.Lock()
		defer b.rwMtx.Unlock()
		close(b.msgs)
	}
}
