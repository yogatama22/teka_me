package screens

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAll(c *fiber.Ctx) error {
	data, err := h.service.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "failed to fetch screens",
			"error":   err.Error(),
		})
	}
	return c.JSON(data)
}
