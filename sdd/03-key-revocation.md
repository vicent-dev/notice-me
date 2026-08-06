# SDD: Task 03 — Key Revocation

## Summary

API keys currently have no way to be deactivated. Once created, they work forever. This change adds a `RevokedAt` field to the `ApiKey` entity, checks in the auth middleware whether a key has been revoked (returning 401 if so), and exports a `RevokeApiKey()` function in the auth package to mark a key as revoked via the repository. Revoked keys remain in the database (not deleted).

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `pkg/auth/entity.go` | modify | Add `RevokedAt *time.Time` field to `ApiKey` struct |
| 2 | `pkg/auth/service.go` | modify | Add `RevokeApiKey(id string, repo Repository[ApiKey]) error` |
| 3 | `app/middleware.go` | modify | After successful key lookup, check `apiKey.RevokedAt`; return 401 if revoked |

## Current State

### 1. `pkg/auth/entity.go`

```go
package auth

import (
	"crypto/sha256"
	"encoding/hex"
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

func HashApiKey(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}

func NewApiKey() (string, *ApiKey) {
	rawUUID := uuid.New().String()
	return rawUUID, &ApiKey{
		ID:           uuid.New(),
		Value:        HashApiKey(rawUUID),
		RequestCount: 0,
	}
}
```

### 2. `pkg/auth/service.go`

```go
package auth

import (
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
```

### 3. `app/middleware.go`

```go
package app

import (
	"errors"
	"net/http"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"
	"github.com/en-vee/alog"
)

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alog.Info("[" + r.Method + "] " + r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (s *server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader := r.Header.Get(auth.API_KEY_HEADER)
		if apiKeyHeader == "" {
			s.writeErrorResponse(w, errors.New("missing "+auth.API_KEY_HEADER+" header"), http.StatusUnauthorized)
			return
		}
		hashedKey := auth.HashApiKey(apiKeyHeader)
		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])
		apiKeysMatch, err := repo.FindBy(repository.Field{Column: "Value", Value: hashedKey})
		if err != nil || len(apiKeysMatch) == 0 {
			s.writeErrorResponse(w, errors.New("invalid API key"), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

## Detailed Changes

### 1. `pkg/auth/entity.go`

**Changes:**
- Add `"time"` import
- Add `RevokedAt *time.Time` field to `ApiKey` struct with gorm column tag

**After (full file):**
```go
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
```

### 2. `pkg/auth/service.go`

**Changes:**
- Add `RevokeApiKey(id string, repo Repository[ApiKey]) error` function
- The function uses `repo.Find(id)` to locate the key, sets `RevokedAt` to current time, and calls `repo.Update(ak)` to persist

**After (full file):**
```go
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
```

### 3. `app/middleware.go`

**Changes:**
- After finding the matching API key, check if `apiKey.RevokedAt != nil`
- If revoked, return 401 with "API key has been revoked" error
- Use the first match from `apiKeysMatch` (the slice)

**After (full file):**
```go
package app

import (
	"errors"
	"net/http"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"

	"github.com/en-vee/alog"
)

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		alog.Info("[" + r.Method + "] " + r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (s *server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyHeader := r.Header.Get(auth.API_KEY_HEADER)
		if apiKeyHeader == "" {
			s.writeErrorResponse(w, errors.New("missing "+auth.API_KEY_HEADER+" header"), http.StatusUnauthorized)
			return
		}

		hashedKey := auth.HashApiKey(apiKeyHeader)

		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

		apiKeysMatch, err := repo.FindBy(repository.Field{Column: "Value", Value: hashedKey})

		if err != nil || len(apiKeysMatch) == 0 {
			s.writeErrorResponse(w, errors.New("invalid API key"), http.StatusUnauthorized)
			return
		}

		// Check if the key has been revoked
		if apiKeysMatch[0].RevokedAt != nil {
			s.writeErrorResponse(w, errors.New("API key has been revoked"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

## Data Flow

### Key Revocation
```
Caller → RevokeApiKey(id, repo)
  ├─ repo.Find(id) → *ApiKey (or error)
  ├─ ak.RevokedAt = &time.Now()
  └─ repo.Update(ak) → persists revocation
```

### Auth Middleware (updated)
```
HTTP Request (X-API-Key: ...)
  ├─ Hash header → hashedKey
  ├─ repo.FindBy(Value=hashedKey) → []*ApiKey
  ├─ Check: len==0? → 401 "invalid API key"
  ├─ Check: RevokedAt != nil? → 401 "API key has been revoked"
  └─ Pass → next.ServeHTTP()
```

## Test Plan

### New test file: `pkg/auth/service_test.go`

| # | Test Case | Description | Input | Expected Output |
|---|-----------|-------------|-------|-----------------|
| 1 | `TestRevokeApiKeySuccess` | Revoke an existing key | Mock repo with one key; call `RevokeApiKey("0", repo)` | No error; `RevokedAt` is set on the entity |
| 2 | `TestRevokeApiKeyNotFound` | Revoke a non-existent key | Mock repo with no matching key | Returns error "API key not found" |
| 3 | `TestGenerateApiKeyHasNilRevokedAt` | Newly generated keys have nil RevokedAt | `NewApiKey()` | Returned entity has `RevokedAt == nil` |

### Tests for middleware:

Add test cases in a new file or update `app/middleware_test.go`:
| # | Test Case | Description | Expected |
|---|-----------|-------------|----------|
| 1 | `TestAuthMiddlewareActiveKey` | Request with valid, active key | Passes (calls next handler) |
| 2 | `TestAuthMiddlewareRevokedKey` | Request with valid but revoked key | Returns 401 "API key has been revoked" |
| 3 | `TestAuthMiddlewareMissingKey` | No X-API-Key header | Returns 401 "missing X-API-Key header" |

## Edge Cases & Failure Modes

### Revoking Already-Revoked Key
- Calling `RevokeApiKey` twice on the same key is safe — it just updates `RevokedAt` again with the current time. No error is returned since the key exists.

### Key Not Found
- `repo.Find(id)` returns an error. `RevokeApiKey` returns `"API key not found"`.

### Nil RevokedAt on New Keys
- `NewApiKey()` explicitly sets `RevokedAt: nil`. The middleware's check `RevokedAt != nil` will not trigger for brand-new keys.

### Database Migration
- Existing keys in the DB will have `RevokedAt = NULL`. The middleware's nil check handles this correctly — they will be treated as active.
- Schema migration: GORM's `AutoMigrate` will add the `revoked_at` column with default null.
