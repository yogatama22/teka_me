package ws

import (
	"time"

	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func RegisterRoutes(router fiber.Router) {
	// WebSocket configuration with keepalive to prevent idle timeout
	wsConfig := websocket.Config{
		EnableCompression: false,
		// Keepalive settings to prevent cloud platform idle timeouts
		HandshakeTimeout: 10 * time.Second,
	}

	// WebSocket per order
	router.Get("/orders/:orderID", websocket.New(WebSocketHandler, wsConfig))

	// Chat per order (Protected by JWT)
	router.Get("/chat/:orderID", middleware.WebSocketJWTProtected(), websocket.New(ChatWebSocketHandler, wsConfig))
}
