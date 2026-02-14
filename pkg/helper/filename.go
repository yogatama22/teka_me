package helper

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"strings"
)

// GenerateRandomFileName bikin nama file acak + ekstensi asli
func GenerateRandomFileName(originalName string) string {
	ext := filepath.Ext(originalName)

	randomBytes := make([]byte, 8) // 16 hex char
	_, _ = rand.Read(randomBytes)

	return fmt.Sprintf("%x%s", randomBytes, ext)
}

// SanitizeFileName bersihin nama folder/user
func SanitizeFileName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name
}
