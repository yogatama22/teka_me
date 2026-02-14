package ws

import (
"teka-api/pkg/middleware"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/websocket/v2"
)

func RegisterRoutes(router fiber.Router) {
// WebSocket configuration to disable compression
wsConfig := websocket.Config{
EnableCompression: false,
}

// WebSocket per order
router.Get("/orders/:orderID", websocket.New(WebSocketHandler, wsConfig))

// Chat per order (Protected by JWT)
router.Get("/chat/:orderID", middleware.WebSocketJWTProtected(), websocket.New(ChatWebSocketHandler, wsConfig))
}
