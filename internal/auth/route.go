package auth

import (
	middleware "teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	// Buat grup API sebagai parent
	api := app.Group("/api")

	// Routes umum di bawah /api
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("teka-api is running 🚀")
	})

	api.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":    "v1.0.1",
			"service":    "teka-api",
			"tanggal":    "13-02-2026",
			"keterangan": "web socket",
		})
	})

	// Auth routes: /api/auth/...
	auth := api.Group("/auth")
	auth.Post("/register", RegisterController)
	auth.Post("/verify-otp", VerifyOtpController)
	auth.Post("/login", LoginController)
	auth.Post("/resend-otp", ResendOtpController)

	// PIN routes: /api/pin/... (protected)
	pin := api.Group("/pin", middleware.JWTProtected())
	pin.Get("/check", CheckPin)
	pin.Post("/create", CreatePin)
	pin.Put("/update", UpdatePin)
	pin.Post("/verify", VerifyPin)

	// Role routes: /api/role/... (protected)
	role := api.Group("/role", middleware.JWTProtected())
	role.Post("/switch-role", SwitchRoleController)
	role.Post("/active-role", GetActiveRole)

	// User routes: /api/user/... (protected)
	user := api.Group("/user", middleware.JWTProtected())
	user.Get("/profile", GetProfile)

	// ✅ REGISTER FCM TOKEN
	user.Post("/fcm", RegisterFCM)
}
