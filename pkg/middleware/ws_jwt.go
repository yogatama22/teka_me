package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
)

// WebSocketJWTProtected validates JWT for WebSocket connections
// Token can be passed via:
// 1. Query parameter: ?token=xxx
// 2. Sec-WebSocket-Protocol header (for browser compatibility)
func WebSocketJWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if this is a WebSocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			var tokenString string

			// Try to get token from query parameter first
			tokenString = c.Query("token")

			// If not in query, try Authorization header
			if tokenString == "" {
				authHeader := c.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
				}
			}

			// If still not found, try Sec-WebSocket-Protocol header
			if tokenString == "" {
				protocols := c.Get("Sec-WebSocket-Protocol")
				parts := strings.Split(protocols, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					// Skip "Bearer" protocol name itself
					if p != "Bearer" && p != "" {
						tokenString = p
						break
					}
				}
			}

			if tokenString == "" {
				return c.Status(fiber.StatusUnauthorized).
					JSON(fiber.Map{"error": "Missing token"})
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

		// If not a WebSocket request, use standard JWT middleware
		return JWTProtected()(c)
	}
}
