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
