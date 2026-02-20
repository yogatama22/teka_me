package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	// 4️⃣ Fiber with proper configuration for production
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		// Disable body limit for file uploads
		BodyLimit: 50 * 1024 * 1024, // 50MB
		// Enable case sensitive routing
		CaseSensitive: true,
		// Enable strict routing
		StrictRouting: false,
		// Server header
		ServerHeader: "TekaPro",
		// App name
		AppName: "TekaPro API",
	})

	// CORS middleware - CRITICAL for WebSocket in production
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // In production, specify your domains
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Sec-WebSocket-Protocol,Sec-WebSocket-Key,Sec-WebSocket-Version,Upgrade,Connection",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length,Content-Type",
		MaxAge:           3600,
	}))

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))

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

	// Realtime routes (WebSocket)
	realtime.RegisterRoutes(app)

	// 🔥 START DISPATCH WORKER
	go dokterService.RunOfferTimeoutWorker(context.Background())

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
