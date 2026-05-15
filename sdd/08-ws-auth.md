# SDD: Task 08 — Authenticate WebSocket

## Summary

The WebSocket endpoint `/ws` currently has no authentication. This change adds API key validation in the WebSocket upgrade handler, checking the `X-API-Key` header first, then falling back to an `apiKey` query parameter. The key is validated using the same SHA-256 hash lookup as the REST middleware.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `app/handler.go` | modify | Add auth check in `wsHandler()` before WebSocket upgrade |

## Current State

### `app/handler.go` — `wsHandler()` function

```go
func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws
	cors := s.c.Server.Cors

	return func(w http.ResponseWriter, r *http.Request) {
		hub.Upgrader.CheckOrigin = func(r *http.Request) bool {
			for _, host := range cors {
				if host == r.Host {
					return true
				}

				if host == "*" {
					return true
				}
			}

			return false
		}

		id := r.URL.Query().Get("id")
		group := r.URL.Query().Get("groupId")

		if id == "" {
			id = hub.AllClientId
		}

		if group == "" {
			group = hub.AllClientGroupId
		}

		conn, err := hub.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		client := hub.NewClient(
			id,
			group,
			ws,
			conn,
			make(chan []byte, 256),
		)

		client.WebsocketService.RegisterClient(client)

		go client.Write()
		go client.Read()
	}
}
```

## Detailed Changes

### 1. `app/handler.go`

**Changes:**
- Add auth validation at the start of the returned handler function
- Read `X-API-Key` from request header; if empty, read `apiKey` from query parameter
- Hash the key value using `auth.HashApiKey()`
- Look up the key in the auth repository
- Check `RevokedAt` for active status
- Return 401 if key is missing, invalid, or revoked

**New imports needed**: `"errors"` (already imported), `"notice-me-server/pkg/auth"` (already imported), `"notice-me-server/pkg/repository"` (already imported).

**Modified section:**
```go
func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws
	cors := s.c.Server.Cors

	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate: check X-API-Key header first, fall back to query param
		apiKeyValue := r.Header.Get(auth.API_KEY_HEADER)
		if apiKeyValue == "" {
			apiKeyValue = r.URL.Query().Get("apiKey")
		}

		if apiKeyValue == "" {
			s.writeErrorResponse(w, errors.New("missing "+auth.API_KEY_HEADER+" header or apiKey query param"), http.StatusUnauthorized)
			return
		}

		// Validate the key using hash lookup
		hashedKey := auth.HashApiKey(apiKeyValue)
		repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])
		apiKeysMatch, err := repo.FindBy(repository.Field{Column: "Value", Value: hashedKey})

		if err != nil || len(apiKeysMatch) == 0 {
			s.writeErrorResponse(w, errors.New("invalid API key"), http.StatusUnauthorized)
			return
		}

		// Check if key is revoked
		if apiKeysMatch[0].RevokedAt != nil {
			s.writeErrorResponse(w, errors.New("API key has been revoked"), http.StatusUnauthorized)
			return
		}

		hub.Upgrader.CheckOrigin = func(r *http.Request) bool {
			for _, host := range cors {
				if host == r.Host {
					return true
				}

				if host == "*" {
					return true
				}
			}

			return false
		}

		id := r.URL.Query().Get("id")
		group := r.URL.Query().Get("groupId")

		if id == "" {
			id = hub.AllClientId
		}

		if group == "" {
			group = hub.AllClientGroupId
		}

		conn, err := hub.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		client := hub.NewClient(
			id,
			group,
			ws,
			conn,
			make(chan []byte, 256),
		)

		client.WebsocketService.RegisterClient(client)

		go client.Write()
		go client.Read()
	}
}
```

## Test Plan

### Update `app/handler_test.go` — `TestWSHandlerSuccess`

The existing WS test needs to be updated to pass an API key. Additionally, add new test cases.

| # | Test Case | Description | Expected |
|---|-----------|-------------|----------|
| 1 | `TestWSHandlerMissingKey` | Connect without X-API-Key header | Returns 401 |
| 2 | `TestWSHandlerInvalidKey` | Connect with invalid X-API-Key | Returns 401 |
| 3 | `TestWSHandlerValidKey` | Connect with valid X-API-Key via header | Upgrade succeeds (101 Switching Protocols) |

The existing test needs to be updated to set up an auth repo and pass a valid key.

## Edge Cases

- **Missing key header and query param**: Returns 401 with descriptive message.
- **Wrong key**: Returns 401 "invalid API key".
- **Revoked key**: Returns 401 "API key has been revoked".
- **Valid key via header**: Standard path — upgrade succeeds.
- **Valid key via query param**: Fallback path — upgrade succeeds. Note: query param keys may be logged by proxies; header is preferred.
- **WebSocket after auth**: All existing WebSocket functionality (id, groupId params, broadcasting) works identically for authenticated clients.
