package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
)

func (s *server) consumeNotificationNotifyHandler() func([]byte) {
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])
	return func(body []byte) {
		notification.NotifyNotification(repo, s.ws, body)
	}
}

func (s *server) consumeNotificationCreateHandler() func([]byte) {
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])
	apiKeyRepo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])
	return func(body []byte) {
		notification.CreateNotification(repo, apiKeyRepo, s.rabbit, s.ws, body)
	}
}
