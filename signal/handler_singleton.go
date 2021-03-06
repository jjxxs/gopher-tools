package signal

import (
	"os"
	"sync"
)

// Holds the singleton
var (
	once     = sync.Once{}
	instance Handler
)

// Handle - Handles a set of signals with the specified function, e.g.
// the cbs-function is called when one of the signals is
// received
func Handle(handler func(sig os.Signal), signals ...os.Signal) {
	once.Do(func() {
		instance = NewHandler()
	})

	instance.Register(handler, signals...)
}

// HandleOneShot - Handles a set of signals with a specified function.
// The function is called exactly once if any of the given signals are received.
func HandleOneShot(handler func(sig os.Signal), signals ...os.Signal) {
	once.Do(func() {
		instance = NewHandler()
	})
	instance.RegisterOneShot(handler, signals...)
}
