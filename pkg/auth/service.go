package auth

import (
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
