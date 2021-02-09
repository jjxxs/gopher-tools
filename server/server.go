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
