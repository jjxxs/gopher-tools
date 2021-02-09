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
	upgrader = DemilitarizedUpgrader(100, false)
)

func TestConnectionSendReceive(t *testing.T) {
	// server
	svrSideMsgStream, svrSideMsgHandler := getMessageStreamWithHandler(nil)
	svrSideConns := serverAcceptConnectionsAt(t, t.Name(), svrSideMsgHandler)

	// client that echoes every message it receives
	_, clientSideMsgHandler := getMessageStreamWithHandler(echoHandler)
	_ = connectToServerAt(t, t.Name(), clientSideMsgHandler)

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
	_, svrSideMsgHandler := getMessageStreamWithHandler(nil)
	svrSideConns := serverAcceptConnectionsAt(t, t.Name(), svrSideMsgHandler)

	// client
	_, clientSideMsgHandler := getMessageStreamWithHandler(nil)
	clientSideConn := connectToServerAt(t, t.Name(), clientSideMsgHandler)

	// wait for client to connect
	svrSideConn := waitForConnectionOrFail(t, svrSideConns, 100*time.Millisecond)

	// server close event
	srvSideCloseEvents, srvSideCloseEventHandler := getCloseEventStream()
	svrSideConn.OnClose(srvSideCloseEventHandler)

	// client close events
	clientSideCloseEvents, clientSideCloseEventHandler := getCloseEventStream()
	clientSideConn.OnClose(clientSideCloseEventHandler)

	// client closes the connection
	clientSideConn.Shutdown()

	// both client and server-side should fire the close-event
	waitForCloseEventOrFail(t, srvSideCloseEvents, 100*time.Millisecond)
	waitForCloseEventOrFail(t, clientSideCloseEvents, 100*time.Millisecond)
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
	srvSideConns := serverAcceptConnectionsAt(t, t.Name(), svrSideMsgHandler)
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
		clientSideConns[i] = connectToServerAt(t, t.Name(), clientSideMsgHandler)
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

func getCloseEventStream() (chan bool, func(Connection, websocket.CloseError)) {
	closeEvents := make(chan bool, 10)
	onClose := func(this Connection, err websocket.CloseError) {
		closeEvents <- true
	}
	return closeEvents, onClose
}

func waitForCloseEventOrFail(t *testing.T, closeEvents chan bool, timeout time.Duration) {
	select {
	case <-closeEvents:
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

func serverAcceptConnectionsAt(t *testing.T, pattern string, onSrvSideMsg func(this Connection, msgType int, data []byte)) chan Connection {
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
		conn := NewConnection(c, onSrvSideMsg)
		inConns <- conn
	})
	return inConns
}

func connectToServerAt(t *testing.T, pattern string, onClientSideMsg func(this Connection, msgType int, data []byte)) Connection {
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d/%s", port, pattern), nil)
	if err != nil {
		t.Fail()
	}
	conn := NewConnection(c, onClientSideMsg)
	return conn
}