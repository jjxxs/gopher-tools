package bus

import "time"

// Type of the Message carried by this Bus - change it to a specific type if necessary.
type Message = interface{}

// A Subscriber can receive/handle messages published on the Bus.
type Subscriber interface {
	HandleMessage(msg Message)
}

// A Bus provides an implementation of a loosely-coupled publish-subscriber
// pattern. Subscriber(s) can subscribe to the Bus and are called whenever a
// Message is Publish(ed) on the Bus.
type Bus interface {
	// Publish a Message on the Bus. The Message will be forwarded to
	// all Subscriber(s).
	Publish(msg Message)

	// Subscribe to the Bus. The given Subscriber will be notified
	// whenever a message is Publish(ed) on the Bus.
	Subscribe(sub Subscriber)

	// Unsubscribe from the Bus. No further Message(s) will be
	// received by given Subscriber.
	Unsubscribe(sub Subscriber)
}

// A WorkerBus uses a queue to buffer message that are passed
// via Bus.Publish. Publishing on a WorkerBus will block, if the
// queue is full or return immediately. A WorkerBus employs go-
// routines to deliver messages to Subscriber(s).
type WorkerBus interface {
	Bus

	// Publishes a message on the Bus. If the queue is full, waits
	// a maximum amount of time before cancelling the operation.
	// Returns true of the message was enqueued, false if not.
	PublishTimeout(msg Message, timeout time.Duration) bool
}
