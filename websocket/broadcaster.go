package websocket

import (
	"sync"
)

// Aggregates Connection to broadcast messages
type Broadcaster interface {
	// Add a connection to the Broadcaster
	Add(conn Connection)
	// Remove a connection from the Broadcaster
	Remove(conn Connection)
	// Broadcast message with MessageType and data
	Broadcast(msgType int, data []byte)
}

type broadcasterImpl struct {
	mtx         *sync.RWMutex
	connections map[string]Connection
}

func NewBroadcaster() Broadcaster {
	return &broadcasterImpl{
		mtx:         &sync.RWMutex{},
		connections: make(map[string]Connection),
	}
}

func (b *broadcasterImpl) Add(conn Connection) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.connections[conn.Conn().RemoteAddr().String()] = conn
}

func (b *broadcasterImpl) Remove(conn Connection) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	delete(b.connections, conn.Conn().RemoteAddr().String())
}

func (b *broadcasterImpl) Broadcast(msgType int, data []byte) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	for _, conn := range b.connections {
		_ = conn.Send(msgType, data)
	}
}
