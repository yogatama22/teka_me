package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// InitMinio aman, bucket coba dipastikan tapi server tetap jalan walau gagal create
func InitMinio() *minio.Client {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	if endpoint == "" || accessKey == "" || secretKey == "" {
		log.Println("⚠️ S3 credentials not set, MinIO disabled")
		return nil
	}

	// hapus protocol kalau ada
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	secure := os.Getenv("S3_USE_SSL") != "false" // default true

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
		Region: os.Getenv("S3_REGION"),
	})
	if err != nil {
		log.Println("⚠️ MinIO init failed:", err)
		return nil
	}

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		log.Println("⚠️ S3_BUCKET not set, skipping bucket check")
		return client
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Println("⚠️ Bucket check failed:", err)
		return client
	}

	if !exists {
		if os.Getenv("S3_ALLOW_CREATE") == "true" {
			fmt.Println("⚠️ Bucket not found, trying to create...")
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
				Region: os.Getenv("S3_REGION"),
			}); err != nil {
				log.Println("⚠️ Failed to create bucket, you must create it manually:", err)
			} else {
				fmt.Println("✅ Bucket successfully created:", bucket)
			}
		} else {
			log.Println("⚠️ Bucket does not exist. Create manually or set S3_ALLOW_CREATE=true")
		}
	} else {
		fmt.Println("✅ MinIO ready, bucket exists:", bucket)
	}

	return client
}
