package firebase

import (
	"context"
	"log"

	"teka-api/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// PushGlobalHandler
// Push notification ke SEMUA FCM token di DB
// Jika DB kosong → pakai dummy token (testing backend only)
type PushGlobalRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

func PushGlobalHandler(c *fiber.Ctx) error {
	if FcmClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "FCM client not initialized"})
	}

	var req PushGlobalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if req.Title == "" || req.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "title and body are required"})
	}

	// 1. Ambil semua token unik dari DB
	var rows []struct {
		FCMToken string `gorm:"column:fcm_token"`
	}

	if err := database.DB.
		Table("user_fcm_tokens").
		Select("DISTINCT fcm_token").
		Find(&rows).Error; err != nil {
		log.Println("❌ Failed to fetch FCM tokens:", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(rows) == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "no tokens found"})
	}

	tokens := make([]string, 0)
	for _, r := range rows {
		if r.FCMToken != "" {
			tokens = append(tokens, r.FCMToken)
		}
	}

	// 2. Kirim secara async supaya respon API cepat
	go func(title, body string, tokens []string, data map[string]string) {
		log.Printf("🚀 Starting global broadcast: %s (%d tokens)", title, len(tokens))
		SendFCMToTokens(
			context.Background(),
			tokens,
			title,
			body,
			data,
		)
		log.Println("✅ Global broadcast finished")
	}(req.Title, req.Body, tokens, req.Data)

	return c.JSON(fiber.Map{
		"message": "broadcast started",
		"tokens":  len(tokens),
	})
}
