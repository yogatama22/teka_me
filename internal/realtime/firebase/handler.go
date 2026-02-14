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
func PushGlobalHandler(c *fiber.Ctx) error {
	// pastikan Firebase sudah di-init
	if FcmClient == nil { // Updated usage
		return c.Status(500).JSON(fiber.Map{
			"error": "FCM client not initialized",
		})
	}

	// ambil semua token dari DB
	var rows []struct {
		FCMToken string `gorm:"column:fcm_token"`
	}

	if err := database.DB.
		Table("user_fcm_tokens").
		Select("fcm_token").
		Find(&rows).Error; err != nil {

		log.Println("❌ Failed to fetch FCM tokens:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// kumpulkan token
	tokens := make([]string, 0)
	for _, r := range rows {
		if r.FCMToken != "" {
			tokens = append(tokens, r.FCMToken)
		}
	}

	// fallback dummy token
	if len(tokens) == 0 {
		log.Println("⚠️ No FCM tokens found, using dummy tokens")
		tokens = []string{
			"DUMMY_FCM_TOKEN_1",
			"DUMMY_FCM_TOKEN_2",
			"DUMMY_FCM_TOKEN_3",
		}
	}

	// kirim menggunakan utility function agar payload konsisten (APNS + Android High Priority)
	results := SendFCMToTokens(
		context.Background(),
		tokens,
		"Test Global Notification",
		"Ini contoh push notification global ke semua user",
		nil,
	)

	success := 0
	failure := 0
	for _, err := range results {
		if err != nil {
			failure++
		} else {
			success++
		}
	}

	log.Printf(
		"✅ FCM push done | success=%d failure=%d total=%d\n",
		success, failure, len(tokens),
	)

	return c.JSON(fiber.Map{
		"success": success,
		"failure": failure,
		"tokens":  len(tokens),
	})
}
