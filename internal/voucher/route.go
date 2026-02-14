package voucher

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"teka-api/pkg/middleware"
)

func Routes(app *fiber.App, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	api := app.Group("/api")

	// USER
	api.Get("/vouchers", middleware.JWTProtected(), handler.GetUserVouchers)

	// ADMIN
	admin := api.Group("/admin", middleware.JWTProtected())
	admin.Get("/vouchers", handler.GetAdminVouchers)
}
