package firebase

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(router fiber.Router) {
	router.Post("/push-global", PushGlobalHandler)
}
