package server

// Handles connections made by a server
type ConnectionHandler interface {
	OnConnect(conn Connection)
	OnDisconnect(conn Connection)
}

// A multiplexConnectionHandler aggregates multiple connection-handlers
// into a single connection-handler
type multiplexConnectionHandler struct {
	connectionHandlers []ConnectionHandler
}

func NewMultiplexConnectionHandler(cs ...ConnectionHandler) ConnectionHandler {
	return &multiplexConnectionHandler{
		connectionHandlers: append(make([]ConnectionHandler, 0), cs...),
	}
}

func (m multiplexConnectionHandler) OnConnect(conn Connection) {
	for _, ch := range m.connectionHandlers {
		ch.OnConnect(conn)
	}
}

func (m multiplexConnectionHandler) OnDisconnect(conn Connection) {
	for _, ch := range m.connectionHandlers {
		ch.OnDisconnect(conn)
	}
}
