package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Invalid Authorization format"})
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Token is empty"})
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "JWT_SECRET is not configured"})
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(
					fiber.StatusUnauthorized,
					"Unexpected signing method",
				)
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Invalid token claims"})
		}

		// REQUIRED CLAIM
		id, ok := claims["id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "Invalid id claim"})
		}

		// inject ke context
		c.Locals("user_id", uint(id))
		c.Locals("nama", claims["nama"])
		c.Locals("email", claims["email"])
		c.Locals("phone", claims["phone"])

		return c.Next()
	}
}
