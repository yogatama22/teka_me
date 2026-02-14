package helper

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

// UploadFileToMinio uploads a file to MinIO and returns object path
func UploadFileToMinio(
	client *minio.Client,
	bucketName string,
	userDir string,
	file *multipart.FileHeader,
) (string, error) {

	if file == nil {
		return "", nil
	}

	objectName := fmt.Sprintf(
		"%s/%s",
		userDir,
		GenerateRandomFileName(file.Filename), // ⬅️ TIDAK TAMBAH EXT LAGI
	)

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = client.PutObject(
		context.Background(),
		bucketName,
		objectName,
		f,
		file.Size,
		minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return "", err
	}

	return objectName, nil
}
