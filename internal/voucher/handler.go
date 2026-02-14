package voucher

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ADMIN
func (h *Handler) GetAdminVouchers(c *fiber.Ctx) error {
	vouchers, err := h.service.GetActiveVouchers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "failed get vouchers",
		})
	}

	return c.JSON(fiber.Map{"data": vouchers})
}

// USER
func (h *Handler) GetUserVouchers(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	vouchers, err := h.service.GetAvailableVouchers(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "failed get vouchers",
		})
	}

	return c.JSON(fiber.Map{"data": vouchers})
}
