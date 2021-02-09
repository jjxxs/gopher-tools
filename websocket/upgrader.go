package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// GetDemilitarizedUpgrader creates an Upgrader that offers no protection against
// cross site request forgery (csrf). Only use within demilitarized zones.
func GetDemilitarizedUpgrader(bufferSize int, compression bool) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:    bufferSize,
		WriteBufferSize:   bufferSize,
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: compression,
	}
}
