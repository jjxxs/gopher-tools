package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// A Server provides means to deliver static files over http and/or
// accept incoming websocket-connections.
type Server interface {
	AddHttpFileServer(pattern string, path string) error
	AddHttpHandler(pattern string, handler func(http.ResponseWriter, *http.Request))
	AddWsHandler(pattern string, handler WsHandler, upgrader *websocket.Upgrader)
	GetUnderlyingServer() *http.Server
	GetUnderlyingServeMux() *http.ServeMux
	ListenAndServe() error
	Exit() error
}

type serverImpl struct {
	run      bool
	addr     string
	serveMux *http.ServeMux
	server   *http.Server
}

// Created a new Server at the specified address.
func NewServer(addr string) Server {
	s := &serverImpl{
		run:      true,
		addr:     addr,
		serveMux: http.NewServeMux(),
		server:   &http.Server{Addr: addr},
	}

	return s
}

// Add a new file-path to be served by the server at the specified pattern
// Returns error of type *PathError if the given path can't be accessed etc.
func (s *serverImpl) AddHttpFileServer(pattern string, path string) error {
	var err error

	if _, err = os.Stat(path); err == nil {
		s.serveMux.Handle(pattern, http.FileServer(http.Dir(path)))
	}

	return err
}

func (s *serverImpl) AddHttpHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.serveMux.HandleFunc(pattern, handler)
}

// Adds a websocket-handler that handles connections for the given pattern. Tries
// to upgrade the initial connection-request with the specified upgrader. If the
// upgrader is nil, a DemilitarizedWebsocketUpgrader will be used which isn't meant
// to be used in production as it does not check for csrf.
func (s *serverImpl) AddWsHandler(pattern string, handler WsHandler, upgrader *websocket.Upgrader) {
	if upgrader == nil {
		s.serveMux.HandleFunc(pattern, s.handleWsRequest(handler, DemilitarizedWebsocketUpgrader()))
	} else {
		s.serveMux.HandleFunc(pattern, s.handleWsRequest(handler, upgrader))
	}
}

// Starts listening for tcp-connections made to the server and handles them
// with the handlers registered via AddHttpFileServer()- and AddWsHandler().
// Handlers can be added before and after calling ListenAndServe()
func (s *serverImpl) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.serveMux)
}

// Shuts down the server. Gives the underlying server a timeout of five seconds
// to successfully closeRequest. If this fails, an error is returned.
func (s *serverImpl) Exit() error {
	s.run = false

	// use context, give server five seconds for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (s *serverImpl) GetUnderlyingServer() *http.Server {
	return s.server
}

func (s *serverImpl) GetUnderlyingServeMux() *http.ServeMux {
	return s.serveMux
}

func (s *serverImpl) handleWsRequest(handler WsHandler, upgrader *websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// if Exit() was called, ignore the request
		if !s.run {
			return
		}

		// try to upgrade the connection
		if conn, err := upgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
		} else {
			handler.Handle(conn)
		}
	}
}

// DemilitarizedWsUpgraderBufferSize is the size of read/write-buffers
var DemilitarizedWsUpgraderBufferSize = 1024

// DemilitarizedWebsocketUpgrader is a upgrader meant to be used in a
// demilitarized context. Offers no protection against cross site request
// forgery (csrf).
func DemilitarizedWebsocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:    DemilitarizedWsUpgraderBufferSize,
		WriteBufferSize:   DemilitarizedWsUpgraderBufferSize,
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: false,
	}
}
