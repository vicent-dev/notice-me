package auth

import (
	"testing"

	repo_mock "notice-me-server/pkg/repository/mock"
)

func TestRevokeApiKeySuccess(t *testing.T) {
	repo := repo_mock.NewRepository[ApiKey]()
	plaintext, ak := NewApiKey()
	_ = plaintext
	_ = repo.Create(ak)

	err := RevokeApiKey("0", repo)
	if err != nil {
		t.Errorf("RevokeApiKey failed: %v", err)
	}

	// Verify the key was revoked
	saved, err := repo.Find("0")
	if err != nil {
		t.Errorf("Failed to find revoked key: %v", err)
	}
	if saved.RevokedAt == nil {
		t.Errorf("Expected RevokedAt to be set after revocation")
	}
}

func TestRevokeApiKeyNotFound(t *testing.T) {
	repo := repo_mock.NewRepository[ApiKey]()

	err := RevokeApiKey("nonexistent", repo)
	if err == nil {
		t.Errorf("Expected error for non-existent key, got nil")
	}
	if err.Error() != "API key not found" {
		t.Errorf("Expected 'API key not found', got: %v", err)
	}
}

func TestNewApiKeyHasNilRevokedAt(t *testing.T) {
	_, ak := NewApiKey()
	if ak.RevokedAt != nil {
		t.Errorf("Expected RevokedAt to be nil for new key, got %v", ak.RevokedAt)
	}
}
