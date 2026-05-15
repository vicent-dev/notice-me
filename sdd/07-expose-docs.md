# SDD: Task 07 — Expose /api/docs from Auth

## Summary

The `/api/docs` endpoint (Swagger UI) is currently registered on the `/api` subrouter which applies the auth middleware, making the docs page require an API key. This change moves `/api/docs` to the main router so it's freely accessible without authentication.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `app/route.go` | modify | Move `/api/docs` from `apiRouter` to main router `s.r` |

## Current State

```go
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

## Detailed Changes

**Changes:**
- Remove `apiRouter.HandleFunc("/docs", s.docsHandler()).Methods("GET")` from the apiRouter block
- Add `s.r.HandleFunc("/api/docs", s.docsHandler()).Methods("GET")` before the apiRouter block so it's on the main router (no auth middleware)

**After:**
```go
func (s *server) routes() {

	s.r.Use(s.loggingMiddleware)

	// websocket connection
	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	// docs (no auth required)
	s.r.HandleFunc("/api/docs", s.docsHandler()).Methods("GET")

	// api route group
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(s.jsonMiddleware)
	apiRouter.Use(s.authMiddleware)

	// auth key management
	apiRouter.HandleFunc("/auth/keys", s.listKeysHandler()).Methods("GET")
	apiRouter.HandleFunc("/auth/keys/{id}", s.revokeKeyHandler()).Methods("DELETE")

	// notifications CRUD
	apiRouter.HandleFunc("/notifications", s.createNotificationHandler()).Methods("POST")
	apiRouter.HandleFunc("/notifications/notify/{id}", s.notifyNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.getNotificationsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.getNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.deleteNotificationHandler()).Methods("DELETE")
}
```

## Test Plan

No new tests needed. The existing handler test for docs is not affected (it tests the handler function directly, not the route registration).

Manual verification:
1. Access `/api/docs` without `X-API-Key` — should return Swagger UI
2. Access other `/api/*` routes without key — should return 401

## Edge Cases

- **Route conflict**: The full path `/api/docs` is registered on `s.r`. The subrouter prefix `/api` matches first, but since `apiRouter` no longer has the `/docs` sub-path, there's no conflict. mux matches the most specific route first.
- **Order matters**: Registering `/api/docs` on `s.r` before the `apiRouter` ensures mux picks the main router route before the subrouter prefix match. Actually, with gorilla/mux, routes registered with longer paths match first, so it works regardless of order. But placing it before the subrouter block reads clearly.
