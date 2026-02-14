package address

import (
	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	// Buat grup API sebagai parent
	api := app.Group("/api")

	// Address routes: /api/addresses/... (protected)
	address := api.Group("/addresses", middleware.JWTProtected())
	address.Get("/", GetUserAddresses)
	address.Post("/create", CreateUserAddress)
	address.Put("/:id", UpdateUserAddress)    // PUT /api/addresses/:id
	address.Delete("/:id", DeleteUserAddress) // DELETE /api/addresses/:id

	// Address Types routes: /api/address-types/... (protected)
	types := api.Group("/address-types", middleware.JWTProtected())
	types.Get("/", GetAddressTypes)
	types.Post("/create", CreateAddressType)
	types.Delete("/:id", DeleteAddressType)
}
