package signal

import (
	"os"
	"os/signal"
	"sync"
)

/*
 * Handler Singleton
 */

var (
	once     = sync.Once{}
	instance Handler
)

func Handle(handler func(), signals ...os.Signal) {
	once.Do(func() {
		instance = NewHandler()
	})

	instance.Handle(handler, signals...)
}

/*
 * Handler
 */

type Handler interface {
	Handle(handler func(), signals ...os.Signal)
	Reset(signals ...os.Signal)
	Exit()
}

type handlerImpl struct {
	mtx     *sync.Mutex
	handler map[os.Signal][]func()
	signals chan os.Signal
	exit    chan bool
}

func NewHandler() Handler {
	h := handlerImpl{
		mtx:     &sync.Mutex{},
		handler: make(map[os.Signal][]func()),
		signals: make(chan os.Signal, 1),
		exit:    make(chan bool, 1),
	}

	go h.listenForSignals()

	return &h
}

func (h *handlerImpl) Handle(handler func(), signals ...os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for _, sig := range signals {
		if _, ok := h.handler[sig]; !ok {
			h.handler[sig] = make([]func(), 0)
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
			handler()
		}
	}
}
