package realtime

import (
	"teka-api/internal/realtime/ws"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/realtime")
	ws.RegisterRoutes(api)
}
