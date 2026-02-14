package main

import (
	"fmt"
	"log"
	"os"

	"teka-api/internal/address"
	"teka-api/internal/auth"
	"teka-api/internal/global_parameter"
	"teka-api/internal/job_category.go"
	"teka-api/internal/job_tarif"
	dokter "teka-api/internal/mitra/dokter"
	"teka-api/internal/realtime"
	"teka-api/internal/realtime/firebase"
	"teka-api/internal/realtime/redis"
	"teka-api/internal/screens"
	"teka-api/internal/voucher"
	"teka-api/pkg/database"
	"teka-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
)

func main() {
	// 1️⃣ Load env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ .env not found, using system env variables")
	}

	// 2️⃣ Connect DB
	db := database.ConnectDB()
	fmt.Println("✅ Database connected")

	var minioClient *minio.Client
	minioClient = utils.InitMinio()
	orderHub := dokter.NewOrderHub()

	// 4️⃣ Fiberx
	app := fiber.New()

	// 5️⃣ Routes tanpa dependency khusus
	auth.AuthRoutes(app)
	app.Static("/assets", "./assets")
	address.RegisterRoutes(app)
	voucher.Routes(app, db)
	job_tarif.Routes(app, db)

	// Screens
	screenRepo := screens.NewRepository(db)
	screenService := screens.NewService(screenRepo)
	screenHandler := screens.NewHandler(screenService)
	screens.RegisterRoutes(app, screenHandler)

	// Dokter (PAKAI MinIO)
	dokterRepo := dokter.NewRepository(db)
	dokterService := dokter.NewService(dokterRepo, orderHub)
	dokterHandler := dokter.NewHandler(*dokterService, minioClient, orderHub)
	dokter.RegisterRoutes(app, dokterHandler)

	// -------------------------------
	// GLOBAL PARAMETER
	// -------------------------------
	global_parameter.Routes(app, db)

	// -------------------------------
	// JOB CATEGORY
	// -------------------------------
	job_category.Routes(app, db)

	// Init Redis & Firebase
	redis.InitRedis()
	firebase.InitFirebase()

	// Realtime routes
	realtime.RegisterRoutes(app)

	// go func() {
	// 	statuses := []string{"OTW", "ACCEPTED", "ON-ROUTE", "FINISHED"}
	// 	for _, s := range statuses {
	// 		realtime.PublishStatus("123", s)
	// 		time.Sleep(2 * time.Second)
	// 	}
	// }()

	// Healthcheck
	app.Get("/kaithheathcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 6️⃣ Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port", port)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
