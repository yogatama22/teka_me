package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		schema = "myschema"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	// ✅ SILENT LOGGER UNTUK MENGHINDARI LOG BERISIK DI TERMINAL
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             2 * time.Second, // Tetap pantau query yang benar-benar lambat
			LogLevel:                  logger.Silent,   // Silent mode: Sembunyikan semua log query sukses
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("❌ Gagal konek ke DB:", err)
	}

	// ✅ CONFIG CONNECTION POOLING
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ✅ SET search_path
	err = db.Exec("SET search_path TO " + schema + ", public").Error
	if err != nil {
		log.Fatal("❌ Gagal set search_path:", err)
	}

	DB = db
	log.Println("🔌 Database connected (Silent Mode)")
	return db
}
