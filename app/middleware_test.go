package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddlewarePassesThrough(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	rr := httptest.NewRecorder()
	loggingMiddleware(inner).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/probe", nil))

	if !called {
		t.Fatal("loggingMiddleware did not invoke the next handler")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusTeapot)
	}
}

func TestJsonMiddlewareSetsContentType(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	jsonMiddleware(inner).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/x", nil))

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
