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
		alog.Info(r.RequestURI)
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

		_, err := repo.FindBy(repository.Field{Column: "Value", Value: apiKeyHeader})

		if err != nil {
			s.writeErrorResponse(w, errors.New("invalid API key"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
