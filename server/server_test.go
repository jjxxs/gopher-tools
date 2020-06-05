package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"strconv"
	"sync"
	"testing"
	"time"
)

/*
 * This test starts the server and registers a websocket-endpoint at 'wsEndpoint'.
 * It will then establish 'concurrentConnections' to the server. After all
 * connections are established, the server will send a message to all connected
 * clients. The clients will then echo back the message. Server and clients will
 * ping-pong the message 2 * 'txsPerConnection' times.
 */

const (
	port                  = 9042
	concurrentConnections = 1000
	txsPerConnection      = 10
	wsEndpoint            = "/ws"
)

/*
 * Tests
 */

func TestConcurrentServerConnections(t *testing.T) {
	// start server
	server := NewServer(fmt.Sprintf(":%d", port))
	defer server.Exit()

	// add connection-handler and give server some time to boot up
	tch := NewTestConnectionHandler()
	server.AddWebsocketHandler(wsEndpoint, tch)
	time.Sleep(500 * time.Millisecond)

	// create concurrent connections to server
	start := time.Now()
	if err := createClientConnections(); err != nil {
		t.Fatal(err)
	}
	durConnections := time.Since(start)
	t.Log("established", concurrentConnections, "connections")
	t.Log("total time to establish connections (s):", durConnections.Seconds())
	t.Log("avg time per connection (ms):", durConnections.Milliseconds()/int64(concurrentConnections))

	// send message to all clients, they will echo it back
	start = time.Now()
	tch.Broadcast()

	totalMessages := 0
	for range tch.Txs {
		totalMessages++

		if totalMessages == concurrentConnections*txsPerConnection {
			break
		}
	}

	elapsed := time.Since(start)

	t.Log("transactions per connection:", txsPerConnection)
	t.Log("total transactions:", concurrentConnections*txsPerConnection)
	t.Log("total time (ms):", elapsed.Milliseconds())
	t.Log("avg time per transaction (µs):", elapsed.Microseconds()/int64(totalMessages))

	// close all connections, give server 50µs per connection
	tch.CloseAll()
	time.Sleep(50 * time.Microsecond * concurrentConnections)

	// there should be zero connections left
	if cs := len(tch.Clients); cs > 0 {
		t.Fatalf("expected zero clients, got %d\n", cs)
	}
}

func createClientConnections() error {
	for i := 0; i < concurrentConnections; i++ {
		// connect
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+strconv.Itoa(port)+wsEndpoint, nil)
		if err != nil {
			return err
		}

		// the client echoes every message it receives
		go func() {
			for {
				t, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}

				err = conn.WriteMessage(t, msg)
				if err != nil {
					break
				}
			}

			_ = conn.Close()
		}()
	}

	return nil
}

/*
 * Test Connection Handler
 */

type testConnectionHandler struct {
	mtx     *sync.Mutex
	Clients map[string]Connection
	Txs     chan bool
}

func NewTestConnectionHandler() *testConnectionHandler {
	return &testConnectionHandler{
		mtx:     &sync.Mutex{},
		Clients: make(map[string]Connection),
		Txs:     make(chan bool, concurrentConnections*txsPerConnection),
	}
}

func (t *testConnectionHandler) OnConnect(c Connection) {
	t.mtx.Lock()
	t.Clients[c.String()] = c
	t.mtx.Unlock()

	// handle the connection
	go func() {
		txs := 0
		for i := range c.GetInput() {
			c.GetOutput() <- i
			t.Txs <- true
			txs++
			if txs >= txsPerConnection {
				break
			}
		}
	}()
}

func (t *testConnectionHandler) OnDisconnect(conn Connection) {
	t.mtx.Lock()
	delete(t.Clients, conn.String())
	t.mtx.Unlock()
}

func (t *testConnectionHandler) Broadcast() {
	type myMessage struct {
		Uuid    string `json:"uuid"`
		Method  string `json:"method"`
		Payload string `json:"payload"`
	}

	t.mtx.Lock()
	for _, v := range t.Clients {
		msg := myMessage{
			Uuid:    "test-test-test-test",
			Method:  "some-method",
			Payload: "some-payload",
		}

		if bytes, err := json.Marshal(&msg); err != nil {
			return
		} else {
			v.GetOutput() <- bytes
		}
	}
	t.mtx.Unlock()
}

func (t *testConnectionHandler) CloseAll() {
	t.mtx.Lock()
	for _, c := range t.Clients {
		c.Close()
	}
	t.mtx.Unlock()
}
