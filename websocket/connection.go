package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents a connected websocket.
type Connection interface {
	// Send message
	Send(msgType int, data []byte) error
	// Access underlying connection
	Conn() *websocket.Conn
	// Shutdown the connection
	Shutdown()
	// Address of remote endpoint (e.g. "192.0.2.1:25" or "[2001:db8::1]:80")
	String() string
}

type connectionImpl struct {
	conn         *websocket.Conn
	shutdownOnce *sync.Once
	onMessage    func(this Connection, msgType int, data []byte)
	onError      func(this Connection, err error)
	onClose      func(this Connection, code int, text string)
}

func NewConnection(conn *websocket.Conn, onMessage func(this Connection, msgType int, data []byte),
	onClose func(this Connection, code int, text string), onError func(this Connection, err error)) Connection {
	c := &connectionImpl{
		conn:         conn,
		shutdownOnce: &sync.Once{},
		onMessage:    onMessage,
		onError:      onError,
		onClose:      onClose,
	}

	go c.readWorker()

	return c
}

func (c *connectionImpl) Send(msgType int, data []byte) error {
	var err error
	if err = c.conn.WriteMessage(msgType, data); err != nil {
		if c.onError != nil {
			c.onError(c, err)
		}
	}
	return err
}

func (c *connectionImpl) Conn() *websocket.Conn {
	return c.conn
}

func (c *connectionImpl) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.sendCloseMessage()
		err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil && c.onError != nil {
			c.onError(c, err)
		}
	})
}

func (c *connectionImpl) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *connectionImpl) readWorker() {
	for {
		msgType, bytes, err := c.conn.ReadMessage()
		if err == nil && c.onMessage != nil {
			c.onMessage(c, msgType, bytes)
		} else {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				c.closeConnection(*closeErr)
			} else if c.onError != nil {
				c.onError(c, err)
			}
			return // once an error was received, the connection is corrupt
		}
	}
}

// tries to gracefully close the connection by sending a close-message
func (c *connectionImpl) sendCloseMessage() {
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := c.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(1*time.Second))
	if err != nil && c.onError != nil {
		c.onError(c, err)
	}
}

func (c *connectionImpl) closeConnection(closeErr websocket.CloseError) {
	err := c.conn.Close()
	if err != nil && c.onError != nil {
		c.onError(c, err)
	}
	if c.onClose != nil {
		c.onClose(c, closeErr.Code, closeErr.Text)
	}
}
