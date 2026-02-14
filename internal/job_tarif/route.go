package job_tarif

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
	jobTarif := api.Group("/job-tarif", middleware.JWTProtected())

	jobTarif.Get("/:id", handler.GetByID)
}
