package bus

import (
	"reflect"
	"sync"
	"time"
)

// Type of the Message carried by this Bus - change it to a specific type if necessary.
type Message = interface{}

// Type of the Subscriber that subscribe to this Bus - change it to a specific type if necessary.
type Subscriber = func(msg Message)

// MsgQueueSize - Size of the buffer for published messages per subscriber
var MsgQueueSize = 1000

// A Bus provides an implementation of a loosely-coupled publish-subscriber
// pattern. Subscriber(s) can subscribe to the Bus and are called whenever a
// Message is Publish(ed) on the Bus.
type Bus interface {
	// Publish a Message on the Bus. This will pass the Message
	// to all Subscriber(s) currently subscribed to this Bus.
	// Publish will block if the message-queue is full. Otherwise
	// it returns immediately.
	Publish(msg Message)

	// Like Publish but will only block for the given duration.
	// Returns true if the message was written to the bus in time,
	// false if not.
	PublishTimeout(msg Message, timeout time.Duration) bool

	// Subscribe to the Bus. The given Subscriber will be called
	// whenever a message is Publish(ed) on the Bus.
	Subscribe(sub Subscriber)

	// Unsubscribe from the Bus. No further Message(s) will be
	// received.
	Unsubscribe(sub Subscriber)
}

type busImpl struct {
	rwMtx *sync.RWMutex
	subs  []Subscriber
	msgs  chan Message
}

func NewBus() Bus {
	bus := &busImpl{
		rwMtx: &sync.RWMutex{},
		subs:  []Subscriber{},
		msgs:  make(chan Message, MsgQueueSize),
	}

	go bus.worker()

	return bus
}

func (b *busImpl) worker() {
	for msg := range b.msgs {
		b.rwMtx.RLock()
		for _, rcv := range b.subs {
			rcv(msg)
		}
		b.rwMtx.RUnlock()
	}
}

func (b *busImpl) Publish(msg Message) {
	b.msgs <- msg
}

func (b *busImpl) PublishTimeout(msg Message, timeout time.Duration) bool {
	select {
	case b.msgs <- msg:
		return true
	case <-time.Tick(timeout):
	}
	return false
}

func (b *busImpl) Subscribe(sub Subscriber) {
	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	b.subs = append(b.subs, sub)
}

func (b *busImpl) Unsubscribe(sub Subscriber) {
	var rcvs []Subscriber
	rcvPtr1 := reflect.ValueOf(sub).Pointer()

	b.rwMtx.Lock()
	defer b.rwMtx.Unlock()
	for _, rcv2 := range b.subs {
		rcvPtr2 := reflect.ValueOf(rcv2).Pointer()
		if rcvPtr1 != rcvPtr2 {
			rcvs = append(rcvs, rcv2)
		}
	}
	b.subs = rcvs
}
