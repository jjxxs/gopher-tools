package server

/*
 * ConnectionHandler - handles connections made by the server
 */

type ConnectionHandler interface {
	OnConnect(conn Connection)
	OnDisconnect(conn Connection)
}

/*
 * multiplexConnectionHandler - combines multiple connection-handlers into one
 */

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
