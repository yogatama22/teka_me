package job_category

import (
	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Routes buat job_category
func Routes(app *fiber.App, db *gorm.DB) {
	// 1️⃣ Buat repo, service, handler
	repo := NewJobCategoryRepository(db)
	service := NewJobCategoryService(repo)
	handler := NewHandler(service)

	// 2️⃣ API group
	api := app.Group("/api")
	jobCategory := api.Group("/job-category", middleware.JWTProtected()) // pake JWT

	// JobCategory routes
	jobCategory.Get("/", handler.GetAll)
	jobCategory.Get("/:id", handler.GetByID)
	jobCategory.Post("/", handler.Create)
	jobCategory.Put("/:id", handler.Update)
	jobCategory.Delete("/:id", handler.Delete)

	// JobSubCategory routes
	jobSub := api.Group("/job-sub-category")
	jobSub.Get("/:id", handler.GetSubByID)                           // GET by sub category ID
	jobSub.Get("/category/:category_id", handler.GetSubByCategoryID) // GET all by job_category_id
	jobSub.Post("/", handler.CreateSub)
	jobSub.Put("/:id", handler.UpdateSub)
	jobSub.Delete("/:id", handler.DeleteSub)
}
