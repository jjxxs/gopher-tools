package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// DemilitarizedUpgrader offers no protection against cross site request
// forgery (csrf). Only use within demilitarized zones.
func DemilitarizedUpgrader(bufferSize int, compression bool) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:    bufferSize,
		WriteBufferSize:   bufferSize,
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: compression,
	}
}
