package server

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

/**
 * Server - can serve websockets and static files via http
 */

type Server interface {
	AddFileHandler(pattern string, path string)
	AddWebsocketHandler(pattern string, handler ConnectionHandler)
	Exit()
}

type serverImpl struct {
	run                    bool
	server                 *http.Server
	wsConnectionHandler    map[string]ConnectionHandler
	wsConnectionHandlerMtx *sync.Mutex
	wsConnections          map[net.Addr]Connection
	wsConnectionsMtx       *sync.Mutex
}

func NewServer(addr string) Server {
	s := &serverImpl{
		run:                    true,
		server:                 &http.Server{Addr: addr},
		wsConnectionHandler:    make(map[string]ConnectionHandler),
		wsConnectionHandlerMtx: &sync.Mutex{},
		wsConnections:          make(map[net.Addr]Connection),
		wsConnectionsMtx:       &sync.Mutex{},
	}

	// start to listen and serve
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Print(err)
		}
	}()

	return s
}

func (s *serverImpl) AddFileHandler(pattern string, path string) {
	http.Handle(pattern, http.FileServer(http.Dir(path)))
}

func (s *serverImpl) AddWebsocketHandler(pattern string, handler ConnectionHandler) {
	if s.wsConnectionHandler == nil {
		return
	}

	s.wsConnectionHandlerMtx.Lock()
	s.wsConnectionHandler[pattern] = handler
	s.wsConnectionHandlerMtx.Unlock()

	http.HandleFunc(pattern, s.handleWsRequestWithPattern(pattern))
}

func (s *serverImpl) Exit() {
	s.run = false

	// close all connected wsConnections
	s.wsConnectionsMtx.Lock()
	for _, c := range s.wsConnections {
		c.Close()
	}
	s.wsConnectionsMtx.Unlock()

	// try to shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *serverImpl) handleWsRequestWithPattern(pattern string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// if Exit() was called, ignore the request
		if !s.run {
			return
		}

		// try to upgrade the connection
		if conn, err := wsUpgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
		} else {
			s.addWsConnection(pattern, conn)
		}
	}
}

func (s *serverImpl) addWsConnection(pattern string, conn *websocket.Conn) {
	var connHandler ConnectionHandler

	// get connection handler for pattern
	s.wsConnectionHandlerMtx.Lock()
	connHandler, _ = s.wsConnectionHandler[pattern]
	s.wsConnectionHandlerMtx.Unlock()
	if connHandler == nil {
		return
	}

	// create new connection
	removeClientOnClose := func() { s.removeWsConnection(pattern, conn) }
	connection := NewBufferedWebsocketConnection(conn, removeClientOnClose)

	// inform the registered connection-handler
	s.wsConnectionHandler[pattern].OnConnect(connection)

	// add connection to map
	s.wsConnectionsMtx.Lock()
	s.wsConnections[conn.RemoteAddr()] = connection
	s.wsConnectionsMtx.Unlock()
}

func (s *serverImpl) removeWsConnection(pattern string, conn *websocket.Conn) {
	var connHandler ConnectionHandler

	// get connection handler for pattern
	s.wsConnectionHandlerMtx.Lock()
	connHandler = s.wsConnectionHandler[pattern]
	s.wsConnectionHandlerMtx.Unlock()
	if connHandler == nil {
		return
	}

	s.wsConnectionsMtx.Lock()
	removeConn := s.wsConnections[conn.RemoteAddr()]
	if removeConn != nil {
		// inform the registered connection-handler
		connHandler.OnDisconnect(removeConn)

		// remove connection from map
		delete(s.wsConnections, conn.RemoteAddr())
	}
	s.wsConnectionsMtx.Unlock()
}
