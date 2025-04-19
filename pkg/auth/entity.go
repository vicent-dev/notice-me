package auth

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const API_KEY_HEADER = "X-API-Key"
const RepositoryKey = "auth"

type ApiKey struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid"`
	Value        string
	RequestCount int
}

func NewApiKey() *ApiKey {
	return &ApiKey{
		ID:           uuid.New(),
		Value:        uuid.New().String(),
		RequestCount: 0,
	}
}
