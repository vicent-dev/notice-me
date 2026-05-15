package auth

import (
	"errors"
	"time"

	"notice-me-server/pkg/repository"
)

// GenerateApiKey creates a new API key, persists the hashed entity to the
// repository, and returns the plaintext key exactly once. The plaintext is
// irrecoverable after this function returns — only the SHA-256 hash is stored.
func GenerateApiKey(repo repository.Repository[ApiKey]) (string, error) {
	plaintext, ak := NewApiKey()

	err := repo.Create(ak)
	if err != nil {
		return "", err
	}

	return plaintext, nil
}

// RevokeApiKey marks an API key as revoked by setting its RevokedAt timestamp
// to the current time. Returns an error if the key is not found.
func RevokeApiKey(id string, repo repository.Repository[ApiKey]) error {
	ak, err := repo.Find(id)
	if err != nil {
		return errors.New("API key not found")
	}

	now := time.Now()
	ak.RevokedAt = &now
	return repo.Update(ak)
}
