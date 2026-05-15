# SDD: Task 05 — Management API Endpoints

## Summary

Add REST endpoints to list and revoke API keys: `GET /api/auth/keys` and `DELETE /api/auth/keys/{id}`. These endpoints are protected by the existing auth middleware. They require adding a `ListApiKeys()` function to the auth service and a new handler file `app/auth_handler.go`.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `pkg/auth/service.go` | modify | Add `ListApiKeys(repo Repository[ApiKey]) ([]ApiKey, error)` that returns keys without exposing the Value field |
| 2 | `app/auth_handler.go` | new | Create handler with `listKeysHandler()` and `revokeKeyHandler()` |
| 3 | `app/route.go` | modify | Register GET `/api/auth/keys` and DELETE `/api/auth/keys/{id}` |

## Current State

### 1. `pkg/auth/service.go`

```go
package auth

import (
	"errors"
	"time"

	"notice-me-server/pkg/repository"
)

func GenerateApiKey(repo repository.Repository[ApiKey]) (string, error) {
	plaintext, ak := NewApiKey()
	err := repo.Create(ak)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

func RevokeApiKey(id string, repo repository.Repository[ApiKey]) error {
	ak, err := repo.Find(id)
	if err != nil {
		return errors.New("API key not found")
	}

	now := time.Now()
	ak.RevokedAt = &now
	return repo.Update(ak)
}
```

### 2. `app/route.go`

```go
package app

func (s *server) routes() {

	s.r.Use(s.loggingMiddleware)

	// websocket connection
	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	// api route group
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(s.jsonMiddleware)
	apiRouter.Use(s.authMiddleware)

	// notifications CRUD
	apiRouter.HandleFunc("/docs", s.docsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.createNotificationHandler()).Methods("POST")
	apiRouter.HandleFunc("/notifications/notify/{id}", s.notifyNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.getNotificationsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.getNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.deleteNotificationHandler()).Methods("DELETE")
}
```

## Detailed Changes

### 1. `pkg/auth/service.go`

**Changes:**
- Add `ListApiKeys(repo Repository[ApiKey]) ([]ApiKey, error)` function
- Uses `repo.FindPaginated` with a large page size to get all keys
- Returns the entities but note: the Value field (hash) is included in the struct but callers should not expose it in responses

**After:**
```go
package auth

import (
	"errors"
	"time"

	"notice-me-server/pkg/repository"
)

// GenerateApiKey creates a new API key, persists the hashed entity to the
// repository, and returns the plaintext key exactly once.
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

// ListApiKeys returns all API keys from the repository.
// The returned keys include the hashed Value field — callers must not expose it.
func ListApiKeys(repo repository.Repository[ApiKey]) ([]ApiKey, error) {
	result, err := repo.FindPaginated(1000, 1)
	if err != nil {
		return nil, err
	}

	keys := make([]ApiKey, len(result.Rows))
	for i, row := range result.Rows {
		keys[i] = *row
	}
	return keys, nil
}
```

### 2. `app/auth_handler.go` (new)

Create a new handler file following the same patterns as `app/handler.go`:

```go
package app

import (
	"net/http"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"

	"github.com/gorilla/mux"
)

// listKeysHandler returns all API keys (without exposing the key value).
func (s *server) listKeysHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

		keys, err := auth.ListApiKeys(repo)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		// Build a safe response that never exposes the key Value
		type keyResponse struct {
			ID           string `json:"id"`
			Active       bool   `json:"active"`
			RequestCount int    `json:"requestCount"`
			CreatedAt    string `json:"createdAt"`
		}

		response := make([]keyResponse, len(keys))
		for i, k := range keys {
			response[i] = keyResponse{
				ID:           k.ID.String(),
				Active:       k.RevokedAt == nil,
				RequestCount: k.RequestCount,
				CreatedAt:    k.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}

		s.writeResponse(w, response)
	}
}

// revokeKeyHandler revokes an API key by ID.
func (s *server) revokeKeyHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

		err := auth.RevokeApiKey(id, repo)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusNotFound)
			return
		}

		s.writeResponse(w, nil)
	}
}
```

### 3. `app/route.go`

**Changes:**
- Add auth key management routes under the existing `apiRouter`
- Register GET `/api/auth/keys` and DELETE `/api/auth/keys/{id}`

**After:**
```go
package app

func (s *server) routes() {

	s.r.Use(s.loggingMiddleware)

	// websocket connection
	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	// api route group
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(s.jsonMiddleware)
	apiRouter.Use(s.authMiddleware)

	// auth key management
	apiRouter.HandleFunc("/auth/keys", s.listKeysHandler()).Methods("GET")
	apiRouter.HandleFunc("/auth/keys/{id}", s.revokeKeyHandler()).Methods("DELETE")

	// notifications CRUD
	apiRouter.HandleFunc("/docs", s.docsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.createNotificationHandler()).Methods("POST")
	apiRouter.HandleFunc("/notifications/notify/{id}", s.notifyNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.getNotificationsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.getNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.deleteNotificationHandler()).Methods("DELETE")
}
```

## Data Flow

```
GET /api/auth/keys
  → authMiddleware (validates key)
  → listKeysHandler()
     → auth.ListApiKeys(repo)
        → repo.FindPaginated(1000, 1)
     → builds safe response without Value field
     → 200 JSON array of {id, active, requestCount, createdAt}

DELETE /api/auth/keys/{id}
  → authMiddleware (validates key)
  → revokeKeyHandler()
     → auth.RevokeApiKey(id, repo)
        → repo.Find(id) → error if not found → 404
        → ak.RevokedAt = now
        → repo.Update(ak)
     → 204 (nil response)
```

## Test Plan

### `pkg/auth/service_test.go` — add:

| # | Test Case | Description | Expected |
|---|-----------|-------------|----------|
| 1 | `TestListApiKeysReturnsKeys` | List keys with keys in repo | Returns expected number of keys |
| 2 | `TestListApiKeysEmpty` | List keys with empty repo | Returns empty slice, no error |

### `app/auth_handler_test.go` — new file:

| # | Test Case | Description | Expected |
|---|-----------|-------------|----------|
| 1 | `TestListKeysHandlerSuccess` | Valid request returns 200 with keys | Response contains keys with safe fields |
| 2 | `TestRevokeKeyHandlerSuccess` | Valid revoke returns 204 | Key is revoked in repo |
| 3 | `TestRevokeKeyHandlerNotFound` | Revoke non-existent key returns 404 | 404 response |

## Edge Cases

- **Empty key list**: `ListApiKeys` returns empty slice, handler returns `[]`.
- **Key not found on revoke**: `RevokeApiKey` returns error, handler returns 404.
- **Value field leakage**: The handler explicitly builds a response DTO that excludes `Value`. Even if `ApiKey` gains new fields, the response struct is explicit.
- **Auth middleware protects endpoints**: Both endpoints sit behind the existing auth middleware, so they require a valid `X-API-Key`.
