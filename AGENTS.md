# Project: Notice-Me API Key Feature

## Project Overview

Go microservice for async frontend notifications via WebSockets.
Uses: gorilla/mux, gorilla/websocket, gorm (MariaDB), RabbitMQ, go 1.23.

## Task List

The API key feature has 8 tasks to reach production readiness:

1. **Fix CLI tool** — `cmd/cli/main.go` calls `NewServer()` which requires DB + RabbitMQ.
   Make key generation work with only a DB connection (no RabbitMQ, no routes, no HTTP).
   Files: `cmd/cli/main.go`, `app/server.go`, `app/cli.go`, `app/db.go`

2. **Hash API keys** — Keys stored as plaintext UUIDs.
   Hash (SHA-256) on storage, show plaintext once at creation, lookup by hash in middleware.
   Files: `pkg/auth/entity.go`, `pkg/auth/service.go`, `app/middleware.go`, `cmd/cli/main.go`

3. **Key revocation** — No way to deactivate a key.
   Add `Active bool` and/or `RevokedAt *time.Time` to `ApiKey`. Check in middleware.
   Add `RevokeApiKey()` service function.
   Files: `pkg/auth/entity.go`, `app/middleware.go`, `pkg/auth/service.go`

4. **Increment RequestCount** — Field exists but never updated.
   Increment in auth middleware after successful key lookup.
   Files: `app/middleware.go`

5. **Management API endpoints** — No way to list/revoke keys via REST.
   Add `GET /api/auth/keys` and `DELETE /api/auth/keys/{id}`.
   Files: `app/route.go`, `app/auth_handler.go` (new), `pkg/auth/service.go`

6. **Makefile CLI target** — No build target for CLI binary.
   Add `build-cli` target.
   Files: `Makefile`

7. **Expose /api/docs from auth** — Swagger docs requires API key.
   Register `/api/docs` on main router (outside auth middleware).
   Files: `app/route.go`

8. **Authenticate WebSocket** — `/ws` has no auth.
   Validate `X-API-Key` header or query param on WebSocket upgrade.
   Files: `app/handler.go`, `app/route.go`

## Project Conventions

- Go 1.23, `gorm.io/gorm` with `gorm.io/driver/mysql`
- Router: `gorilla/mux`
- WebSocket: `gorilla/websocket`
- Testing: standard `testing` package, mock repositories in `pkg/repository/mock/`
- Error responses use `s.writeErrorResponse(w, err, http.StatusXxx)`
- Success responses use `s.writeResponse(w, data)`
- Config is YAML loaded from embedded files via `static/embed.go`
- Repositories follow `repository.Repository[T]` interface
- Tests use `repo_mock.NewRepository[T]()` and `mock.NewRabbitMock()`
- Consumers bound in `app/consumer.go`, handlers in `app/handler.go`
- Always run `go build ./...` and `go test ./... -cover` after changes
