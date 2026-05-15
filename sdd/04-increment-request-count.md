# SDD: Task 04 — Increment RequestCount

## Summary

The `RequestCount` field exists on `ApiKey` but is never incremented. This change updates the auth middleware to increment `RequestCount` after a successful key lookup and persist the updated count via `repo.Update()`.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `app/middleware.go` | modify | Increment `RequestCount` on matched key after successful lookup |

## Current State

### `app/middleware.go`

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

## Detailed Changes

### 1. `app/middleware.go`

**Changes:**
- After the revoked check passes (the key is active), increment `apiKeysMatch[0].RequestCount`
- Call `repo.Update(apiKeysMatch[0])` to persist the updated count
- Ignore the error from Update (best-effort — we don't want to fail the request if the count increment fails)

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

		// Increment request count (best-effort, do not fail on error)
		apiKeysMatch[0].RequestCount++
		_ = repo.Update(apiKeysMatch[0])

		next.ServeHTTP(w, r)
	})
}
```

## Test Plan

Add test case in `app/middleware_test.go`:

| # | Test Case | Description | Expected |
|---|-----------|-------------|----------|
| 1 | `TestAuthMiddlewareIncrementsRequestCount` | After a successful auth, RequestCount increases by 1 | RequestCount goes from 0 to 1 |

## Edge Cases

- **Update failure**: If `repo.Update` fails (e.g., DB connection lost), the error is silently ignored with `_ = repo.Update(...)`. The request still proceeds. This is best-effort to avoid failing valid requests due to a non-critical counter.
- **Concurrent requests**: Two simultaneous requests for the same key may race on the update. GORM's `Update` does a full save, so the last write wins. The count will be eventually consistent. This is acceptable for a usage counter.
