package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents a connected websocket.
type Connection interface {
	// Send message with MessageType and data
	Send(msgType int, data []byte) error
	OnError(h func(err error))
	OnClose(h func(err websocket.CloseError))
	// Access underlying connection
	Conn() *websocket.Conn
	// Closes the connection
	Shutdown()
	String() string
}

type connectionImpl struct {
	conn         *websocket.Conn
	shutdownOnce *sync.Once
	onMessage    func(this Connection, msgType int, data []byte)
	onError      func(err error)
	onClose      func(err websocket.CloseError)
}

func NewConnection(conn *websocket.Conn, onMessage func(this Connection, msgType int, data []byte)) Connection {
	c := &connectionImpl{
		conn:         conn,
		shutdownOnce: &sync.Once{},
		onMessage:    onMessage,
		onError:      nil,
		onClose:      nil,
	}

	go c.readWorker()

	return c
}

func (c *connectionImpl) Send(msgType int, data []byte) error {
	var err error
	if err = c.conn.WriteMessage(msgType, data); err != nil {
		if c.onError != nil {
			c.onError(err)
		}
	}
	return err
}

func (c *connectionImpl) OnError(h func(err error)) {
	c.onError = h
}

func (c *connectionImpl) OnClose(h func(err websocket.CloseError)) {
	c.onClose = h
}

func (c *connectionImpl) Conn() *websocket.Conn {
	return c.conn
}

func (c *connectionImpl) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.sendCloseMessage()
		err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil && c.onError != nil {
			c.onError(err)
		}
	})
}

func (c *connectionImpl) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *connectionImpl) readWorker() {
	for {
		msgType, bytes, err := c.conn.ReadMessage()
		if err == nil {
			c.onMessage(c, msgType, bytes)
		} else {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				c.closeConnection(*closeErr)
			} else if c.onError != nil {
				c.onError(err)
			}
			return // once an error was received, the connection is corrupt
		}
	}
}

// tries to gracefully close the connection by sending a close-message
func (c *connectionImpl) sendCloseMessage() {
	t := time.Now().Add(1 * time.Second)
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := c.conn.WriteControl(websocket.CloseMessage, closeMsg, t)
	if err != nil && c.onError != nil {
		c.onError(err)
	}
}

func (c *connectionImpl) closeConnection(closeErr websocket.CloseError) {
	err := c.conn.Close()
	if err != nil && c.onError != nil {
		c.onError(err)
	}
	if c.onClose != nil {
		c.onClose(closeErr)
	}
}
