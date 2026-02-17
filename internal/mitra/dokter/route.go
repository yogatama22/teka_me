package dokter

import (
	"teka-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func RegisterRoutes(app *fiber.App, h *Handler) {
	// Parent API
	api := app.Group("/api")

	// -------------------------------
	// CUSTOMER (PASIEN)
	// -------------------------------
	customer := api.Group("/customer", middleware.JWTProtected())
	// Search dokter
	customer.Post("/doctors/search", h.SearchDoctor)
	// List / history service order
	customer.Get("/service-orders", h.GetMyServiceOrders)
	// ✅ CURRENT / ACTIVE ORDER (INI YANG BARU)
	customer.Get("/current-order", h.GetCustomerCurrentOrder)
	// Cancel request (sebelum accepted)
	customer.Post("/service-orders/:id/cancel", h.CancelOrder)
	// Complete order (by user) - NEW
	customer.Post("/service-orders/:id/complete", h.CompleteOrderUser)
	// Rate doctor
	customer.Post("/rate", h.RateDoctor)

	// -------------------------------
	// DOKTER / MITRA
	// -------------------------------
	dokter := api.Group("/dokter", middleware.JWTProtected())
	// Dokter profile
	dokter.Post("/register", h.RegisterMitra)
	dokter.Get("/me", h.GetMyMitraProfile)
	// Offer flow
	dokter.Get("/current-offer", h.GetCurrentOffer)
	dokter.Post("/offers/:id/accept", h.AcceptOffer)
	// ✅ CURRENT / ACTIVE SERVICE ORDER (INI YANG BARU)
	dokter.Get("/current-order", h.GetMitraCurrentOrder)
	// Complete service
	dokter.Post("/transaction/:id/complete", h.CompleteOrder)
	// Rating stats (Monthly breakdown)
	dokter.Get("/rating-summary", h.GetRatingSummary)
	dokter.Get("/rating-history", h.GetRatingHistory)
	// Order stats
	dokter.Get("/order-summary", h.GetOrderSummary)
	dokter.Get("/order-history", h.GetOrderHistory)

	// 🔥 UPDATE STATUS (OTW / ARRIVED / COMPLETED)
	dokter.Put("/service-orders/:id/status", h.UpdateServiceOrderStatus)

	// ===============================
	// WEBSOCKET (TIDAK DI DALAM JWT GROUP)
	// ===============================
	ws := app.Group("/ws")

	// realtime status order
	ws.Get(
		"/orders/:order_id",
		websocket.New(h.OrderStatusWS),
	)

}
