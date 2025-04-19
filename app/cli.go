package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"
)

func (s *server) GenerateApiKeyHandler() (*auth.ApiKey, error) {
	repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

	return auth.GenerateApiKey(repo)
}
