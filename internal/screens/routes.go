package screens

import (
	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, h *Handler) {
	api := app.Group("/api")
	screens := api.Group("/screens", middleware.JWTProtected())
	screens.Get("/", h.GetAll)
}
