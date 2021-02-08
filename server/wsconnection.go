package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WsConnectionEvent string

const (
	OnReadError  WsConnectionEvent = "onReadError"
	OnWriteError WsConnectionEvent = "onWriteError"
)

// A WsConnection represents a websocket-connection made by a server.
// It provides high level-access to the connection. The Input()- and
// Output()-functions can be used to access the input/output of the
// connection. The channels they provide are closed when the connection
// closes.
type WsConnection interface {
	// Provides channel for incoming data, read data received by this ws here
	Input() chan []byte
	// Provides channel for outgoing data, write data to be sent here
	Output() chan []byte
	// Sets handler that is called whenever a WsConnectionEvent happens
	SetEventHandler(func(event WsConnectionEvent))
	// Sets a handler that is called when the connection is closed. The handler will be called exactly once.
	SetCloseHandler(onClose func())
	// Access to the underlying conn
	UnderlyingConn() *websocket.Conn
	// Closes the connection
	Close()
	// String representation of the remove host of this connection (e.g. ip:port)
	String() string
}

// A BufferedWsConnection wraps a underlying websocket-connection with read-
// and write-buffers. It will buffer read messages in an input-buffer which
// can be read from through the Input()-method. It will buffer to-be-sent messages in
// the an output-buffer which can be written to through the Output()-method.
type BufferedWsConnection struct {
	conn                      *websocket.Conn
	closeOnce                 *sync.Once
	closeHandler              func()
	eventHandler              func(event WsConnectionEvent)
	stopRead, stopWrite       chan bool
	inputBuffer, outputBuffer chan []byte
}

// BufferedWebsocketConnectionBufferLength is the length of messages to be buffered by the connection
var BufferedWebsocketConnectionBufferLength = 100

// BufferedWsConnection provides high-level access to an underlying websocket-connection.
func NewBufferedWsConnection(conn *websocket.Conn) WsConnection {
	c := &BufferedWsConnection{
		conn:         conn,
		closeOnce:    &sync.Once{},
		stopRead:     make(chan bool, 1),
		stopWrite:    make(chan bool, 1),
		inputBuffer:  make(chan []byte, BufferedWebsocketConnectionBufferLength),
		outputBuffer: make(chan []byte, BufferedWebsocketConnectionBufferLength),
	}

	conn.CloseHandler()

	go c.tryRead()
	go c.tryWrite()

	return c
}

func (c *BufferedWsConnection) Input() chan []byte {
	return c.inputBuffer
}

func (c *BufferedWsConnection) Output() chan []byte {
	return c.outputBuffer
}

func (c *BufferedWsConnection) SetEventHandler(eventHandler func(event WsConnectionEvent)) {
	c.eventHandler = eventHandler
}

func (c *BufferedWsConnection) SetCloseHandler(closeHandler func()) {
	c.closeHandler = closeHandler
}

func (c *BufferedWsConnection) UnderlyingConn() *websocket.Conn {
	return c.conn
}

func (c *BufferedWsConnection) Close() {
	c.closeOnce.Do(func() {
		c.stopRead <- true
		c.stopWrite <- true
		_ = c.conn.Close()
		if c.closeHandler != nil {
			c.closeHandler()
		}
	})
}

func (c *BufferedWsConnection) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *BufferedWsConnection) tryRead() {
	defer func() { c.Close() }()

	for {
		select {
		case <-c.stopRead:
			return
		default:
			if _, bytes, err := c.conn.ReadMessage(); err != nil {
				if c.eventHandler != nil {
					c.eventHandler(OnReadError)
				}
				return
			} else {
				c.inputBuffer <- bytes
			}
		}
	}
}

func (c *BufferedWsConnection) tryWrite() {
	defer func() { c.Close() }()

	for {
		select {
		case <-c.stopWrite:
			return
		case msg := <-c.outputBuffer:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				if c.eventHandler != nil {
					c.eventHandler(OnWriteError)
				}
				return
			}
		}
	}
}
