package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// This test starts the server and registers a websocket-endpoint at 'wsEndpoint'.
// It will then establish n='concurrentConnections' to the server. After all
// connections are established, the server will send a message to all connected
// clients. The clients will then echo back the message. Server and clients will
// ping-pong the message m='txsPerConnection' times to simulate a query consisting
// of a request and a response.

const (
	port                  = 9042  // port to use for this test
	concurrentConnections = 1000  // concurrent connections
	txsPerConnection      = 100   // transactions per connection
	wsEndpoint            = "/ws" // servers websocket-endpoint
)

func TestConcurrentServerConnections(t *testing.T) {
	// start server at localhost:port
	server := NewServer(fmt.Sprintf(":%d", port))

	// add websocket-handler to server and start listening for connection
	cond := &sync.Cond{L: &sync.Mutex{}}
	testWsHandler := NewTestWsHandler(cond)
	server.AddWsHandler(wsEndpoint, testWsHandler, nil)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // give ListenAndServe() go-routine some time to execute

	// create concurrent connections to server and output statistics
	concurrentConnectionsWithStats(t)

	// send message to all clients
	broadcastMsgWithStats(t, testWsHandler)

	// wait until all transactions are done
	cond.Broadcast()
	waitForTransactionsWithStats(t, testWsHandler)

	// exiting the server should return no error
	if err := server.Exit(); err != nil {
		t.Fatal(err)
	}
}

func broadcastMsgWithStats(t *testing.T, wsHandler BroadcastWsHandler) {
	type testMessage struct {
		Uuid    string `json:"uuid"`
		Method  string `json:"method"`
		Payload string `json:"payload"`
	}
	msg := testMessage{
		Uuid:    "test-test-test-test",
		Method:  "some-method",
		Payload: "some-payload",
	}
	bytes, _ := json.Marshal(&msg)

	start := time.Now()
	wsHandler.Broadcast(bytes)
	since := time.Since(start)
	t.Log("sent", len(bytes), "bytes as broadcast:")
	t.Log("time total (ms):", since.Milliseconds())
	t.Log("time avg (µs):", since.Microseconds()/concurrentConnections)
}

func waitForTransactionsWithStats(t *testing.T, handler *testWsHandler) {
	totalMessages := 0
	msHistogram := make([]int64, 100)
	msOverflowHistogram := make([]int64, 9)
	min, max := int64(math.MaxInt64), int64(math.MinInt64)
	start := time.Now()
	for tx := range handler.Txs {
		totalMessages++
		if totalMessages >= concurrentConnections*txsPerConnection {
			break
		}

		if min > tx {
			min = tx
		} else if max < tx {
			max = tx
		}

		bucket := tx / 1000.0
		if bucket < 100 {
			msHistogram[bucket]++
		} else if bucket >= 900 {
			msOverflowHistogram[8]++
		} else if bucket >= 800 {
			msOverflowHistogram[7]++
		} else if bucket >= 700 {
			msOverflowHistogram[6]++
		} else if bucket >= 600 {
			msOverflowHistogram[5]++
		} else if bucket >= 500 {
			msOverflowHistogram[4]++
		} else if bucket >= 400 {
			msOverflowHistogram[3]++
		} else if bucket >= 300 {
			msOverflowHistogram[2]++
		} else if bucket >= 200 {
			msOverflowHistogram[1]++
		} else if bucket >= 100 {
			msOverflowHistogram[0]++
		}
	}
	since := time.Since(start)
	t.Log("finished", totalMessages, "transactions:")
	t.Log("tx time total (ms):", since.Milliseconds())
	t.Log("tx time avg (µs)", since.Microseconds()/int64(totalMessages))
	t.Log("tx time max (µs)", max)
	t.Log("tx time min (µs)", min)
	t.Log("tx histogram (0-99ms):")
	t.Logf("%8s %8s", "ms", "count")
	for i, v := range msHistogram {
		t.Logf("%8d %8d", i, v)
	}
	t.Log("overflow [n-m):")
	for i, v := range msOverflowHistogram {
		t.Logf("%8s %8d", fmt.Sprintf("%d-%d", (i+1)*100, (i+2)*100), v)
	}
	since := time.Since(start)
	t.Log("finished", concurrentConnections*txsPerConnection, "transactions:")
	t.Log("tx time total (ms):", since.Milliseconds())
	t.Log("tx time avg (µs)", since.Microseconds()/int64(txsPerConnection*concurrentConnections))
	t.Log("tx time max (µs)", max)
	t.Log("tx time min (µs)", min)
}

func concurrentConnectionsWithStats(t *testing.T) {
	// make connections
	connDelays := make([]int64, concurrentConnections)
	for i := 0; i < concurrentConnections; i++ {
		if delay, err := connectWithEchoClient(); err != nil {
			t.Fatal(err)
		} else {
			connDelays[i] = delay
		}
	}

	// output statistics
	min, max, sum := int64(math.MaxInt64), int64(math.MinInt64), int64(0)
	for _, t := range connDelays {
		if min > t {
			min = t
		} else if max < t {
			max = t
		}
		sum += t
	}
	t.Log("established", concurrentConnections, "connections to server in:")
	t.Log("conn time total (ms):", sum)
	t.Log("conn time avg (µs):", sum/concurrentConnections)
	t.Log("conn time max (µs):", max)
	t.Log("conn time min (µs):", min)
}

// creates a connection to the specified port and endpoint.
// the connection then echoes every message that is received
// until an error occurs. measures how long it takes to connect
// and returns the duration in milliseconds. the returned error
// is always nil unless the client failed to connect to the server.
func connectWithEchoClient() (int64, error) {

	// connect, measure how long it takes
	start := time.Now()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+strconv.Itoa(port)+wsEndpoint, nil)
	dur := time.Since(start)
	if err != nil {
		return dur.Microseconds(), err
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

	return dur.Microseconds(), nil
}

// testWsHandler starts a go-routine for every connection
// which reads from the connection and echoes the received
// data back to the server
type testWsHandler struct {
	cond    *sync.Cond
	mtx     *sync.RWMutex
	clients map[string]WsConnection
	Txs     chan int64
}

func NewTestWsHandler(cond *sync.Cond) *testWsHandler {
	return &testWsHandler{
		cond:    cond,
		mtx:     &sync.RWMutex{},
		clients: make(map[string]WsConnection),
		Txs:     make(chan int64, concurrentConnections*txsPerConnection),
	}
}

func (h *testWsHandler) Handle(conn *websocket.Conn) {
	c := NewBufferedWsConnection(conn)

	h.mtx.Lock()
	h.clients[conn.RemoteAddr().String()] = c
	h.mtx.Unlock()

	// read n=txsPerConnection messages and echo them back to the server
	go func() {
		h.cond.L.Lock()
		h.cond.Wait()
		h.cond.L.Unlock()
		for txs := 0; txs < txsPerConnection; txs++ {
			start := time.Now()
			in := <-c.Input()
			c.Output() <- in
			h.Txs <- time.Since(start).Microseconds()
		}
	}()
}

func (h *testWsHandler) Broadcast(msg []byte) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	for _, c := range h.clients {
		c.Output() <- msg
	}
}

func (h *testWsHandler) Close() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	for _, v := range h.clients {
		v.Close()
	}
	h.clients = make(map[string]WsConnection)
}
