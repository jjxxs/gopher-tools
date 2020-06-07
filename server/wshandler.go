package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Handles incoming websocket-connections
type WsHandler interface {
	// Handles a new connection
	Handle(conn *websocket.Conn)
}

// Handles incoming connections and provides means to broadcast
// to and close all received connections
type BroadcastWsHandler interface {
	// Handles a new connection
	Handle(conn *websocket.Conn)
	// Broadcasts a message to all connections
	Broadcast(msg []byte)
	// Closes all connections
	Close()
}

// A MultiplexWsHandler aggregates multiple connection-handlers
// into a single connection-handle. This can be used when multiple
// handlers should be informed about a new websocket-connection.
type MultiplexWsHandler struct {
	handlers []WsHandler
}

func NewMultiplexWsHandler(cs ...WsHandler) WsHandler {
	return &MultiplexWsHandler{
		handlers: append(make([]WsHandler, 0), cs...),
	}
}

func (m *MultiplexWsHandler) Handle(conn *websocket.Conn) {
	for _, ch := range m.handlers {
		ch.Handle(conn)
	}
}

// A BroadcastWsHandlerImpl is the default implementation for the
// BroadcastWsHandler-Interface.
type BroadcastWsHandlerImpl struct {
	factory func(conn *websocket.Conn) WsConnection
	mtx     *sync.RWMutex
	conns   map[string]WsConnection
}

func NewBroadcastWsHandler(factory func(conn *websocket.Conn) WsConnection) BroadcastWsHandler {
	return &BroadcastWsHandlerImpl{
		factory: factory,
		mtx:     &sync.RWMutex{},
		conns:   make(map[string]WsConnection),
	}
}

func (w *BroadcastWsHandlerImpl) Handle(conn *websocket.Conn) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.conns[conn.RemoteAddr().String()] = w.factory(conn)
}

func (w *BroadcastWsHandlerImpl) Broadcast(msg []byte) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	for _, c := range w.conns {
		c.Output() <- msg
	}
}

func (w BroadcastWsHandlerImpl) Close() {
	w.mtx.Lock()
	w.mtx.RUnlock()
	for _, c := range w.conns {
		c.Close()
	}
	w.conns = make(map[string]WsConnection)
}
