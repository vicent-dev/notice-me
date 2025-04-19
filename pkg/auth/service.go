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
