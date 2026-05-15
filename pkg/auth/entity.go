package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const API_KEY_HEADER = "X-API-Key"
const RepositoryKey = "auth"

type ApiKey struct {
	gorm.Model
	ID           uuid.UUID  `gorm:"type:uuid"`
	Value        string
	RequestCount int
	RevokedAt    *time.Time `gorm:"default:null"`
}

// HashApiKey computes the SHA-256 digest of plaintext and returns it as a
// lowercase hex-encoded string. This is a one-way function — the original
// value cannot be recovered from the hash.
func HashApiKey(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}

// NewApiKey generates a new API key pair: a plaintext UUID that is shown to
// the caller exactly once, and an ApiKey entity whose Value field stores the
// SHA-256 hash of that UUID. Only the hashed entity should be persisted.
func NewApiKey() (string, *ApiKey) {
	rawUUID := uuid.New().String()
	return rawUUID, &ApiKey{
		ID:           uuid.New(),
		Value:        HashApiKey(rawUUID),
		RequestCount: 0,
		RevokedAt:    nil,
	}
}
