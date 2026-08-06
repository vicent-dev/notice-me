# SDD: Task 2 — Hash API Keys

## Summary

API keys are currently stored as raw UUID plaintext in the database. This is a security vulnerability: if the database is compromised, all API keys are immediately exposed. This task implements SHA-256 hashing of API keys at rest.

The change introduces a `HashApiKey()` function in the `pkg/auth` package that computes `sha256(plaintext) -> hex string`. `NewApiKey()` is modified to generate a raw UUID, hash it before storing, and return both the plaintext (shown exactly once at creation time) and the hashed entity. The auth middleware hashes the incoming `X-API-Key` header before performing the DB lookup. The CLI tool outputs the plaintext key at creation and discards it afterward — no code path can retrieve a plaintext key after initial creation.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `pkg/auth/entity.go` | modify | Add `HashApiKey()` function; change `NewApiKey()` to return `(plaintext string, apiKey *ApiKey)`; hash `Value` before persisting |
| 2 | `pkg/auth/service.go` | modify | Change `GenerateApiKey()` to return `(string, error)` — returns the plaintext key, entity stores the hash |
| 3 | `app/middleware.go` | modify | In `authMiddleware`, hash the `X-API-Key` header via `auth.HashApiKey()` before the repository lookup |
| 4 | `app/cli.go` | modify | Change `GenerateApiKeyCLI()` return type from `(*auth.ApiKey, error)` to `(string, error)` |
| 5 | `cmd/cli/main.go` | modify | Print the returned plaintext string directly instead of `apiKey.Value` |

## Current State

### 1. `pkg/auth/entity.go` (current)

```go
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
```

### 2. `pkg/auth/service.go` (current)

```go
package auth

import (
	"notice-me-server/pkg/repository"
)

func GenerateApiKey(repo repository.Repository[ApiKey]) (*ApiKey, error) {
	ak := NewApiKey()

	err := repo.Create(ak)

	if err != nil {
		return nil, err
	}

	return ak, nil
}
```

### 3. `app/middleware.go` (current)

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

		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

		apiKeysMatch, err := repo.FindBy(repository.Field{Column: "Value", Value: apiKeyHeader})

		if err != nil || len(apiKeysMatch) == 0 {
			s.writeErrorResponse(w, errors.New("invalid API key"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

### 4. `app/cli.go` (current)

```go
package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"

	"gorm.io/gorm"
)

// GenerateApiKeyCLI generates an API key using only a *gorm.DB connection.
// It creates an auth repository, generates a key, persists it, and returns it.
// This function does not require RabbitMQ, HTTP routes, or any other server infrastructure.
func GenerateApiKeyCLI(db *gorm.DB) (*auth.ApiKey, error) {
	repo := repository.NewGorm[auth.ApiKey](db)
	return auth.GenerateApiKey(repo)
}
```

### 5. `cmd/cli/main.go` (current)

```go
package main

import (
	"fmt"
	"notice-me-server/app"
	"notice-me-server/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	db, err := app.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	apiKey, err := app.GenerateApiKeyCLI(db)
	if err != nil {
		fmt.Printf("Error generating API key: %v\n", err)
		return
	}

	fmt.Println(apiKey.Value)
}
```

## Detailed Changes

### 1. `pkg/auth/entity.go`

**Change:** Add imports `crypto/sha256` and `encoding/hex`. Add `HashApiKey(plaintext string) string` function. Modify `NewApiKey()` to return `(string, *ApiKey)` — the first return value is the plaintext UUID shown once at creation; the second is the entity with `Value` set to the SHA-256 hex hash of the plaintext.

**Full file after changes:**

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
	}
}
```

### 2. `pkg/auth/service.go`

**Change:** Update `GenerateApiKey` to use the new `NewApiKey()` signature. It now returns the plaintext key (shown once at creation) instead of the entity. The entity (with `Value = hash`) is persisted.

**Full file after changes:**

```go
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
```

### 3. `app/middleware.go`

**Change:** In the `authMiddleware` function, hash the incoming `X-API-Key` header value using `auth.HashApiKey()` before performing the repository `FindBy` lookup. This ensures the DB query compares against the stored hash rather than the raw plaintext.

**Full file after changes:**

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

### 4. `app/cli.go`

**Change:** Update `GenerateApiKeyCLI` to return `(string, error)` — the plaintext key — instead of `(*auth.ApiKey, error)`. This reflects the updated `auth.GenerateApiKey` signature.

**Full file after changes:**

```go
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
```

### 5. `cmd/cli/main.go`

**Change:** The variable `apiKey` now holds a `string` (the plaintext key) instead of `*auth.ApiKey`. Print it directly with `fmt.Println(apiKey)` instead of `fmt.Println(apiKey.Value)`.

**Full file after changes:**

```go
package main

import (
	"fmt"
	"notice-me-server/app"
	"notice-me-server/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	db, err := app.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	plaintext, err := app.GenerateApiKeyCLI(db)
	if err != nil {
		fmt.Printf("Error generating API key: %v\n", err)
		return
	}

	fmt.Println(plaintext)
}
```

## Data Flow

### Key Generation (CLI)

```
CLI main()
  │
  ├─ config.LoadConfig()
  ├─ app.InitDB(cfg)               → *gorm.DB
  │
  └─ app.GenerateApiKeyCLI(db)
       │
       └─ pkg/auth.GenerateApiKey(repo)
            │
            ├─ uuid.New().String()           → rawUUID = "a1b2c3d4-..."
            │
            ├─ sha256.Sum256([]byte(rawUUID)) → [32]byte
            ├─ hex.EncodeToString(hash[:])    → hashHex = "e3b0c442..."
            │
            ├─ ApiKey{Value: hashHex}         ← entity with hashed value
            │
            ├─ repo.Create(&entity)           ← persists hash only
            │
            └─ return rawUUID, nil            ← plaintext returned to CLI

CLI prints rawUUID  (e.g. "a1b2c3d4-...")
    → plaintext discarded after output
    → attacker who gains DB access sees only "e3b0c442..."
```

### Key Lookup (Middleware)

```
HTTP Request (Header: X-API-Key: a1b2c3d4-...)
  │
  └─ authMiddleware handler
       │
       ├─ r.Header.Get("X-API-Key")       → "a1b2c3d4-..."
       │
       ├─ auth.HashApiKey(plaintext)      → "e3b0c442..."
       │
       └─ repo.FindBy(Field{Column: "Value", Value: hashHex})
            │
            ├─ DB query: SELECT * FROM api_keys WHERE value = 'e3b0c442...'
            │
            └─ if found  → next.ServeHTTP(w, r)
               if not    → 401 Unauthorized
```

## Test Plan

### Existing Tests to Update

- **`app/handler_test.go` — `initialiseMocks()`**: The mock setup does not currently create an auth repository. While no existing test exercises the auth middleware directly (tests invoke handler functions bypassing middleware), adding `auth.RepositoryKey` to the mock map would prevent future panics. This is **optional** for this task — the middleware will only be exercised when the server runs through the router with middleware chains.

### New Test Cases to Add

All to be placed in a new file `pkg/auth/service_test.go`:

| Test Case | Description | Input | Expected Output |
|-----------|-------------|-------|-----------------|
| `TestHashApiKeyDeterministic` | Same input always produces same hash | `HashApiKey("foo")` called twice | Both calls return identical 64-char hex string |
| `TestHashApiKeyLength` | SHA-256 hex output is always 64 chars | `HashApiKey("")`, `HashApiKey(uuid.New().String())` | Len=64 for any input |
| `TestHashApiKeyDifferentInputsDiffer` | Different inputs produce different hashes | `HashApiKey("a")` vs `HashApiKey("b")` | The two outputs are not equal |
| `TestNewApiKeyReturnsDifferentKeys` | Each `NewApiKey()` call produces unique plaintext + unique hash | Call `NewApiKey()` twice | `plaintext1 != plaintext2`, `entity1.Value != entity2.Value` |
| `TestNewApiKeyEntityHasHash` | The entity's `Value` is the hash of the plaintext | `plaintext, entity := NewApiKey()` | `entity.Value == HashApiKey(plaintext)` |
| `TestGenerateApiKeyPersistsAndReturnsPlaintext` | `GenerateApiKey` stores hashed value in repo and returns plaintext | Mock repo; call `GenerateApiKey(mockRepo)` | Returned plaintext is a valid UUID; `mockRepo` contains one entity where `Value == HashApiKey(plaintext)` |
| `TestGenerateApiKeyRepoError` | Error propagation when repo fails | Mock repo that returns error on `Create` | Returns `("", error)` |

**Test helper utilities needed:**

```go
// In pkg/auth/service_test.go:
func TestHashApiKeyDeterministic(t *testing.T) {
    h1 := HashApiKey("test-key-123")
    h2 := HashApiKey("test-key-123")
    if h1 != h2 {
        t.Errorf("HashApiKey is not deterministic: %s != %s", h1, h2)
    }
}

func TestHashApiKeyLength(t *testing.T) {
    h := HashApiKey("")
    if len(h) != 64 {
        t.Errorf("HashApiKey('') length = %d, want 64", len(h))
    }
}

func TestNewApiKeyEntityHasHash(t *testing.T) {
    plaintext, ak := NewApiKey()
    expectedHash := HashApiKey(plaintext)
    if ak.Value != expectedHash {
        t.Errorf("Entity Value = %s, want hash of plaintext = %s", ak.Value, expectedHash)
    }
}
```

## Edge Cases & Failure Modes

### Empty / Missing `X-API-Key` Header
- The middleware **already checks** for an empty header before calling `HashApiKey`, returning `401 Unauthorized`. `HashApiKey("")` would produce a valid SHA-256 hash of the empty string, but this path is never reached due to the guard clause.

### Hash Collisions
- SHA-256 produces 256-bit digests. The probability of a collision among the expected number of API keys (thousands, not billions) is negligible (< 10⁻⁶⁰). No special handling is required.

### Database Compromise
- An attacker with read access to the `api_keys` table sees only 64-character hex strings. Reversing SHA-256 is computationally infeasible. The original UUIDs have 122 bits of randomness (UUID v4), making brute-force impractical.

### Concurrent Generation of the Same UUID
- The probability of `uuid.New()` generating a duplicate across concurrent goroutines is astronomically low (UUID v4 uses 122 random bits). Even in the event of a collision, the `repo.Create` call would fail with a duplicate-key / primary-key error, which `GenerateApiKey` propagates to the caller.

### Old Plaintext Keys in Database
- If the database already contains plaintext keys (from before this migration), the middleware will hash the incoming header and fail to match stored plaintext values. A **data migration** is required to hash existing keys in the DB. This is **out of scope** for this task but should be noted in the deployment plan.

### Repository FindBy Matching on Hash
- The field value is a 64-character hex string. GORM's `Where` clause performs exact matching, so a full hash match is required. No partial/tolerant matching is needed or supported.

### API Key Value Stored in Logs
- The CLI prints the plaintext to stdout. If stdout is captured in a log file, the plaintext is logged at creation time. This is an operational concern that the user must manage (e.g., redirect CLI output to `/dev/null` after capture). The code itself has no codepath for re-printing or logging the plaintext after creation.
