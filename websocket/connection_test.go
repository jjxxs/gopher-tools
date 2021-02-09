package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const port = 9042

var (
	server   *http.Server
	muxer    = &http.ServeMux{}
	upgrader = GetDemilitarizedUpgrader(100, false)
)

func TestConnectionSendReceive(t *testing.T) {
	// server
	svrSideMsgStream, svrSideMsgHandler := getMessageStreamWithHandler(nil)
	svrSideConns := serverAcceptConnectAt(t, t.Name(), svrSideMsgHandler, nil, nil)

	// client that echoes every message it receives
	_, clientSideMsgHandler := getMessageStreamWithHandler(echoHandler)
	_ = clientConnectToServerAt(t, t.Name(), clientSideMsgHandler, nil, nil)

	// wait for client to connect
	svrSideConn := waitForConnectionOrFail(t, svrSideConns, 100*time.Millisecond)

	// send message from server to client
	err := svrSideConn.Send(websocket.TextMessage, []byte("This is a test!"))
	if err != nil {
		t.Fail()
	}
	waitForMessageOrFail(t, svrSideMsgStream, 100*time.Millisecond)
}

func TestConnectionClientCloses(t *testing.T) {
	// server
	srvCloseStream, srvCloseHandler := getCloseEventStream()
	svrSideConns := serverAcceptConnectAt(t, t.Name(), nil, srvCloseHandler, nil)

	// client
	clientCloseStream, clientCloseHandler := getCloseEventStream()
	clientSideConn := clientConnectToServerAt(t, t.Name(), nil, clientCloseHandler, nil)

	// wait for client to connect
	_ = waitForConnectionOrFail(t, svrSideConns, 100*time.Millisecond)

	// client closes connection
	clientSideConn.Shutdown()

	// both client and server-side should fire the close-event
	waitForCloseEventOrFail(t, srvCloseStream, 100*time.Millisecond, 1000, "")
	waitForCloseEventOrFail(t, clientCloseStream, 100*time.Millisecond, 1000, "")
}

func TestConcurrentConnections(t *testing.T) {
	const concurrentConnections = 100
	const messagesPerConnection = 10
	svrSideConns := make([]Connection, concurrentConnections)
	clientSideConns := make([]Connection, concurrentConnections)
	svrSideMsgsReceived := 0
	clientSideMsgsReceived := 0
	svrSideDone := &sync.WaitGroup{}
	svrSideDone.Add(1)
	clientSideDone := &sync.WaitGroup{}
	clientSideDone.Add(1)
	established := &sync.WaitGroup{}
	established.Add(1)

	// server
	svrSideMsgStream, svrSideMsgHandler := getMessageStreamWithHandler(echoHandler)
	srvSideConns := serverAcceptConnectAt(t, t.Name(), svrSideMsgHandler, nil, nil)
	go func() {
		for i := 0; i < concurrentConnections; i++ {
			svrSideConns[i] = <-srvSideConns
		}
		established.Done()
	}()

	// server-message counter
	go func() {
		for range svrSideMsgStream {
			svrSideMsgsReceived++
			if svrSideMsgsReceived == concurrentConnections*messagesPerConnection {
				svrSideDone.Done()
				break
			}
		}
	}()

	// client
	clientSideMsgStream, clientSideMsgHandler := getMessageStreamWithHandler(echoHandler)
	for i := 0; i < concurrentConnections; i++ {
		clientSideConns[i] = clientConnectToServerAt(t, t.Name(), clientSideMsgHandler, nil, nil)
	}

	// client-message counter
	go func() {
		for range clientSideMsgStream {
			clientSideMsgsReceived++
			if clientSideMsgsReceived == concurrentConnections*messagesPerConnection {
				clientSideDone.Done()
				break
			}
		}
	}()

	// wait until all are connected
	established.Wait()

	// send message that round-trips messagesPerConnection times per connection
	s := time.Now()
	for i := 0; i < concurrentConnections; i++ {
		if err := svrSideConns[i].Send(websocket.TextMessage, []byte("This is a test!")); err != nil {
			t.Fail()
		}
	}
	clientSideDone.Wait()
	svrSideDone.Wait()
	d := time.Since(s)
	t.Logf("\nconcurrent connections:\t\t%d\n"+
		"round-trips per connection:\t%d\n"+
		"total messages exchanged:\t%d\n"+
		"total time:\t\t\t\t\t%s\n"+
		"avg time per round-trip:\t%s\n"+
		"avg time per message:\t\t%s", concurrentConnections, messagesPerConnection,
		2*messagesPerConnection*concurrentConnections, d.String(),
		time.Duration(d.Nanoseconds()/(messagesPerConnection*concurrentConnections)).String(),
		time.Duration(d.Nanoseconds()/(2*messagesPerConnection*concurrentConnections)).String())

	if svrSideMsgsReceived != clientSideMsgsReceived {
		t.Fail()
	}
}

type message struct {
	conn    Connection
	msgType int
	data    []byte
}

type closeEvent struct {
	code int
	text string
}

func getCloseEventStream() (chan closeEvent, func(Connection, int, string)) {
	closeEvents := make(chan closeEvent, 10)
	onClose := func(this Connection, code int, text string) {
		closeEvents <- closeEvent{code, text}
	}
	return closeEvents, onClose
}

func waitForCloseEventOrFail(t *testing.T, closeEvents chan closeEvent, timeout time.Duration, code int, text string) {
	select {
	case ce := <-closeEvents:
		if ce.code != code || ce.text != text {
			t.Fail()
		}
	case <-time.After(timeout):
		t.Fatal()
	}
}

func waitForMessageOrFail(t *testing.T, msgs chan message, timeout time.Duration) (msg message) {
	select {
	case msg = <-msgs:
	case <-time.After(timeout):
		t.Fatal()
	}
	return msg
}

func waitForConnectionOrFail(t *testing.T, conns chan Connection, timeout time.Duration) (conn Connection) {
	select {
	case conn = <-conns:
	case <-time.After(timeout):
		t.Fatal()
	}
	return conn
}

func echoHandler(this Connection, msgType int, data []byte) {
	_ = this.Send(msgType, data)
}

func getMessageStreamWithHandler(handler func(Connection, int, []byte)) (chan message, func(Connection, int, []byte)) {
	msgs := make(chan message, 10000)
	onMsg := func(conn Connection, msgType int, data []byte) {
		if handler != nil {
			handler(conn, msgType, data)
		}
		msgs <- message{conn, msgType, data}
	}
	return msgs, onMsg
}

func serverAcceptConnectAt(t *testing.T, pattern string, onMessage func(this Connection, msgType int, data []byte),
	onClose func(this Connection, code int, text string), onError func(this Connection, err error)) chan Connection {
	if server == nil { // the first time this is called we need to start the server
		server = &http.Server{Addr: fmt.Sprintf(":%d", port)}
		go func() {
			err := http.ListenAndServe(server.Addr, muxer)
			if err != nil {
				t.Log(err)
			}
		}()
		time.Sleep(100 * time.Millisecond) // wait for server to come up
	}

	inConns := make(chan Connection, 10000)
	muxer.HandleFunc("/"+pattern, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Log(err)
		}
		conn := NewConnection(c, onMessage, onClose, onError)
		inConns <- conn
	})
	return inConns
}

func clientConnectToServerAt(t *testing.T, pattern string, onMessage func(this Connection, msgType int, data []byte),
	onClose func(this Connection, code int, text string), onError func(this Connection, err error)) Connection {
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d/%s", port, pattern), nil)
	if err != nil {
		t.Fail()
	}
	conn := NewConnection(c, onMessage, onClose, onError)
	return conn
}
