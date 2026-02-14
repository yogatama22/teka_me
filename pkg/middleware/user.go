package middleware

import "github.com/gofiber/fiber/v2"

func UserID(c *fiber.Ctx) (uint, error) {
	id := c.Locals("user_id")
	if id == nil {
		return 0, fiber.ErrUnauthorized
	}

	userID, ok := id.(uint)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}

	return userID, nil
}
