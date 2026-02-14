package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal konek ke DB:", err)
	}

	// ✅ SET search_path DENGAN CARA BENAR
	err = db.Exec("SET search_path TO " + schema + ", public").Error
	if err != nil {
		log.Fatal("❌ Gagal set search_path:", err)
	}

	// 🔍 DEBUG (opsional tapi disarankan)
	var sp string
	db.Raw("SHOW search_path").Scan(&sp)
	log.Println("✅ search_path aktif:", sp)

	DB = db
	log.Println("🔌 Database connected")
	return db
}
