package global_parameter

import (
	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Routes(app *fiber.App, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	api := app.Group("/api")
	params := api.Group("/global-parameters", middleware.JWTProtected()) // ✅ semua route di sini butuh JWT

	// CRUD routes
	params.Get("", handler.GetAll)        // GET /api/global-parameters
	params.Get("/:id", handler.GetByID)   // GET /api/global-parameters/:id
	params.Post("", handler.Create)       // POST /api/global-parameters
	params.Put("/:id", handler.Update)    // PUT /api/global-parameters/:id
	params.Delete("/:id", handler.Delete) // DELETE /api/global-parameters/:id
}
