package server

import (
	"context"
	"net/http"
	"os"
	"time"
)

func AddHttpFileRoot(pattern string, path string, mux *http.ServeMux) error {
	var err error
	if _, err = os.Stat(path); err == nil {
		mux.Handle(pattern, http.FileServer(http.Dir(path)))
	}
	return err
}

func ShutdownWithTimeout(timeout time.Duration, server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

// Adds a websocket-handler that handles connections for the given pattern. Tries
// to upgrade the initial connection-request with the specified upgrader. If the
// upgrader is nil, a DemilitarizedWebsocketUpgrader will be used which isn't meant
// to be used in production as it does not check for csrf.
/*func (s *serverImpl) AddWsHandler(pattern string, handler websocket2.WsHandler, upgrader *websocket.Upgrader) {
	if upgrader == nil {
		s.serveMux.HandleFunc(pattern, s.handleWsRequest(handler, DemilitarizedWebsocketUpgrader()))
	} else {
		s.serveMux.HandleFunc(pattern, s.handleWsRequest(handler, upgrader))
	}
}*/

/*func (s *serverImpl) handleWsRequest(handler websocket2.WsHandler, upgrader *websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// if Shutdown() was called, ignore the request
		if !s.run {
			return
		}

		// try to upgrade the connection
		if conn, err := upgrader.Upgrade(w, r, nil); err != nil {
			log.Println(err)
		} else {
			handler.Add(conn)
		}
	}
}*/
