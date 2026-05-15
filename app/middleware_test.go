package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/config"
	repo_mock "notice-me-server/pkg/repository/mock"
)

func TestAuthMiddlewareMissingKey(t *testing.T) {
	s := &server{
		repositories: make(map[string]interface{}),
		c:            &config.Config{},
	}

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing key, got %d", rr.Code)
	}
}

func TestAuthMiddlewareActiveKey(t *testing.T) {
	repo := repo_mock.NewRepository[auth.ApiKey]()
	plaintext, ak := auth.NewApiKey()
	_ = plaintext
	_ = repo.Create(ak)

	repositories := make(map[string]interface{})
	repositories[auth.RepositoryKey] = repo

	s := &server{
		repositories: repositories,
		c:            &config.Config{},
	}

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(auth.API_KEY_HEADER, plaintext)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for active key, got %d", rr.Code)
	}
}

func TestAuthMiddlewareRevokedKey(t *testing.T) {
	repo := repo_mock.NewRepository[auth.ApiKey]()
	plaintext, ak := auth.NewApiKey()
	_ = plaintext
	_ = repo.Create(ak)

	// Revoke the key
	err := auth.RevokeApiKey("0", repo)
	if err != nil {
		t.Fatal(err)
	}

	repositories := make(map[string]interface{})
	repositories[auth.RepositoryKey] = repo

	s := &server{
		repositories: repositories,
		c:            &config.Config{},
	}

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(auth.API_KEY_HEADER, plaintext)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for revoked key, got %d", rr.Code)
	}
}

func TestAuthMiddlewareIncrementsRequestCount(t *testing.T) {
	repo := repo_mock.NewRepository[auth.ApiKey]()
	plaintext, ak := auth.NewApiKey()
	_ = repo.Create(ak)

	repositories := make(map[string]interface{})
	repositories[auth.RepositoryKey] = repo

	s := &server{
		repositories: repositories,
		c:            &config.Config{},
	}

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(auth.API_KEY_HEADER, plaintext)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify RequestCount was incremented
	key, _ := repo.Find("0")
	if key.RequestCount != 1 {
		t.Errorf("Expected RequestCount=1, got %d", key.RequestCount)
	}
}
