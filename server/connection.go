package server

import (
	"github.com/gorilla/websocket"
	"sync"
)

// A Connection represents a connection made by a server, e.g. a
// websocket-connection
type Connection interface {
	GetInput() chan []byte
	GetOutput() chan []byte
	Close()
	String() string
}

const (
	// Count of messages to buffer
	BufferedWebsocketConnectionBufferLength = 100
)

type bufferedWebsocketConnection struct {
	conn         *websocket.Conn
	inputBuffer  chan []byte
	outputBuffer chan []byte
	close        *sync.WaitGroup
	closeOnce    *sync.Once
	stopRead     chan bool
	stopWrite    chan bool
}

// A BufferedWebsocketConnection provides a buffered websocket-connection
func NewBufferedWebsocketConnection(conn *websocket.Conn, onClose func()) Connection {
	c := &bufferedWebsocketConnection{
		conn:         conn,
		close:        &sync.WaitGroup{},
		closeOnce:    &sync.Once{},
		stopRead:     make(chan bool, 1),
		stopWrite:    make(chan bool, 1),
		inputBuffer:  make(chan []byte, BufferedWebsocketConnectionBufferLength),
		outputBuffer: make(chan []byte, BufferedWebsocketConnectionBufferLength),
	}

	// start reading/writing
	go c.tryRead()
	go c.tryWrite()

	c.close.Add(1)

	go func() {
		c.close.Wait() // block until read/write routines exit because connection closed
		onClose()      // call the provided onClose-function
	}()

	return c
}

func (c *bufferedWebsocketConnection) GetInput() chan []byte {
	return c.inputBuffer
}

func (c *bufferedWebsocketConnection) GetOutput() chan []byte {
	return c.outputBuffer
}

func (c *bufferedWebsocketConnection) Close() {
	c.closeOnce.Do(func() {
		c.stopRead <- true
		c.stopWrite <- true
		_ = c.conn.Close()
		c.close.Done()
	})
}

func (c *bufferedWebsocketConnection) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *bufferedWebsocketConnection) tryRead() {
	defer func() { c.Close() }()

	for {
		select {
		case <-c.stopRead:
			return
		default:
			if _, bytes, err := c.conn.ReadMessage(); err != nil {
				return
			} else {
				c.inputBuffer <- bytes
			}
		}
	}
}

func (c *bufferedWebsocketConnection) tryWrite() {
	defer func() { c.Close() }()

	for {
		select {
		case <-c.stopWrite:
			return
		case msg := <-c.outputBuffer:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}
