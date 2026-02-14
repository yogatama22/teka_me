package redis

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var Rdb *redis.Client

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASSWORD")

	if addr == "" {
		log.Println("⚠️ REDIS_ADDR not set, Redis disabled")
		return
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	if err := Rdb.Ping(ctx).Err(); err != nil {
		log.Println("⚠️ Cannot connect to Redis:", err)
		return // ❗ JANGAN MATIKAN SERVER
	}

	log.Println("✅ Redis connected")
}
