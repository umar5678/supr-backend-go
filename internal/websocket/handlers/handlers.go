package handlers

import (
	"github.com/umar5678/go-backend/internal/websocket"
)

func RegisterAllHandlers(manager *websocket.Manager) {

	RegisterRideHandlers(manager)
}
