package signal

import (
	"os"
	"os/signal"
	"sync"
)

// Holds the singleton
var (
	once     = sync.Once{}
	instance Handler
)

// Handles a set of signals with the specified function, e.g.
// the handler-function is called when one of the signals is
// received
func Handle(handler func(sig os.Signal), signals ...os.Signal) {
	once.Do(func() {
		instance = NewHandler()
	})

	instance.Handle(handler, signals...)
}

// A Handler provides means to register functions that are called
// when the application receives a specified set of signals
type Handler interface {
	Handle(handler func(sig os.Signal), signals ...os.Signal)
	Reset(signals ...os.Signal)
	Exit()
}

type handlerImpl struct {
	mtx     *sync.Mutex
	handler map[os.Signal][]func(sig os.Signal)
	signals chan os.Signal
	exit    chan bool
}

func NewHandler() Handler {
	h := handlerImpl{
		mtx:     &sync.Mutex{},
		handler: make(map[os.Signal][]func(sig os.Signal)),
		signals: make(chan os.Signal, 1),
		exit:    make(chan bool, 1),
	}

	go h.listenForSignals()

	return &h
}

func (h *handlerImpl) Handle(handler func(sig os.Signal), signals ...os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for _, sig := range signals {
		if _, ok := h.handler[sig]; !ok {
			h.handler[sig] = make([]func(sig os.Signal), 0)
			signal.Notify(h.signals, sig)
		}
		h.handler[sig] = append(h.handler[sig], handler)
	}
}

func (h *handlerImpl) Reset(signals ...os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for _, sig := range signals {
		delete(h.handler, sig)
		signal.Reset(sig)
	}
}

func (h *handlerImpl) Exit() {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	// reset all signals
	for k := range h.handler {
		signal.Reset(k)
	}

	// stop listening for signals
	h.exit <- true
}

func (h *handlerImpl) listenForSignals() {
	for {
		select {
		case sig := <-h.signals:
			h.handleSignal(sig)
		case <-h.exit:
			return
		}
	}
}

func (h *handlerImpl) handleSignal(sig os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if _, ok := h.handler[sig]; ok {
		for _, handler := range h.handler[sig] {
			handler(sig)
		}
	}
}
