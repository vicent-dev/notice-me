package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"

	"gorm.io/gorm"
)

// GenerateApiKeyCLI generates an API key using only a *gorm.DB connection.
// It creates an auth repository, generates a key, persists it, and returns
// the plaintext key string exactly once. The stored key is SHA-256 hashed.
// This function does not require RabbitMQ, HTTP routes, or any other server
// infrastructure.
func GenerateApiKeyCLI(db *gorm.DB) (string, error) {
	repo := repository.NewGorm[auth.ApiKey](db)
	return auth.GenerateApiKey(repo)
}
