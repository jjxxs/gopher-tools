package signal

import (
	"os"
	"os/signal"
	"reflect"
	"sync"
)

// Handler provides functionality to register callbacks
// that are called when the application receives a given
// os.Signal.
type Handler interface {
	// Register a callback that is called when any of the
	// provided os.Signal(s) is received.
	Register(cb func(sig os.Signal), signals ...os.Signal)

	// RegisterOneShot registers a callback that is called only
	// once and then removed from the Handler.
	RegisterOneShot(cb func(sig os.Signal), signals ...os.Signal)

	// Unregister a previously registered callback.
	Unregister(cb func(sig os.Signal))

	// Exit will make the Handler stop listening for signals.
	// Once Exit was called, the Handler is useless.
	Exit()
}

type handlerImpl struct {
	mtx     *sync.Mutex
	cbs     map[os.Signal][]func(sig os.Signal)
	signals chan os.Signal
}

func NewHandler() Handler {
	h := handlerImpl{
		mtx:     &sync.Mutex{},
		cbs:     make(map[os.Signal][]func(sig os.Signal)),
		signals: make(chan os.Signal, 10),
	}

	go h.listenForSignals()

	return &h
}

func (h *handlerImpl) Register(cb func(sig os.Signal), signals ...os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	for _, sig := range signals {
		if _, ok := h.cbs[sig]; !ok {
			signal.Notify(h.signals, sig)
		}
		h.cbs[sig] = append(h.cbs[sig], cb)
	}
}

func (h *handlerImpl) RegisterOneShot(cb func(sig os.Signal), signals ...os.Signal) {
	oneShot := &sync.Once{}
	var oneShotFunc func(sig os.Signal)
	oneShotFunc = func(sig os.Signal) {
		oneShot.Do(func() {
			cb(sig)
			h.Unregister(oneShotFunc)
		})
	}
	h.Register(oneShotFunc, signals...)
}

func (h *handlerImpl) Unregister(cb func(os.Signal)) {
	ptr1 := reflect.ValueOf(cb).Pointer()
	h.mtx.Lock()
	defer h.mtx.Unlock()
	for sig, cbs := range h.cbs {
		newCbs := make([]func(os.Signal), 0)
		for _, cb := range cbs {
			ptr2 := reflect.ValueOf(cb).Pointer()
			if ptr1 != ptr2 {
				newCbs = append(newCbs, cb)
			}
		}
		h.cbs[sig] = newCbs
		if len(newCbs) == 0 { // if there are no cbs, stop listening for the signal
			signal.Reset(sig)
		}
	}
}

func (h *handlerImpl) Exit() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	for k := range h.cbs {
		signal.Reset(k) // stop listening for signals
	}
	close(h.signals)
}

func (h *handlerImpl) listenForSignals() {
	for sig := range h.signals {
		h.handleSignal(sig)
	}
}

func (h *handlerImpl) handleSignal(sig os.Signal) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	if cbs, ok := h.cbs[sig]; ok {
		for _, cb := range cbs {
			cb(sig)
		}
	}
}
