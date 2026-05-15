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
