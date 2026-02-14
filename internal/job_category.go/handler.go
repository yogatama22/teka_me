package job_category

import (
	"strconv"

	"teka-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service JobCategoryService
}

func NewHandler(service JobCategoryService) *Handler {
	return &Handler{service}
}

// GET /job-category
func (h *Handler) GetAll(c *fiber.Ctx) error {
	data, err := h.service.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve job categories",
		})
	}
	return c.JSON(data)
}

// GET /job-category/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	category, err := h.service.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job category not found",
		})
	}

	return c.JSON(category)
}

// POST /job-category
func (h *Handler) Create(c *fiber.Ctx) error {
	var category models.JobCategory
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.service.Create(&category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create job category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Successfully created",
		"data":    category,
	})
}

// PUT /job-category/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	var category models.JobCategory
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category.ID = uint(id)
	if err := h.service.Update(&category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully updated",
		"data":    category,
	})
}

// DELETE /job-category/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	if err := h.service.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete job category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully deleted",
	})
}

// --- JobSubCategory ---
func (h *Handler) GetSubByID(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	sub, err := h.service.GetSubByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job sub-category not found"})
	}
	return c.JSON(sub)
}

func (h *Handler) GetSubByCategoryID(c *fiber.Ctx) error {
	catID, _ := strconv.ParseUint(c.Params("category_id"), 10, 64)
	subs, err := h.service.GetSubAllByCategoryID(uint(catID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve job sub-categories"})
	}
	return c.JSON(subs)
}

func (h *Handler) CreateSub(c *fiber.Ctx) error {
	var sub models.JobSubCategory
	if err := c.BodyParser(&sub); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.service.CreateSub(&sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create job sub-category"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Successfully created", "data": sub})
}

func (h *Handler) UpdateSub(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	existingSub, err := h.service.GetSubByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job sub-category not found"})
	}

	var input struct {
		SubCategoryName string `json:"sub_category_name"`
		Description     string `json:"description"`
		Active          bool   `json:"active"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Assign pointer values
	existingSub.SubCategoryName = input.SubCategoryName
	existingSub.Description = &input.Description
	existingSub.Active = &input.Active

	if err := h.service.UpdateSub(existingSub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update job sub-category"})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully updated",
		"data":    existingSub,
	})
}

func (h *Handler) DeleteSub(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.service.DeleteSub(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete job sub-category"})
	}
	return c.JSON(fiber.Map{"message": "Successfully deleted"})
}
