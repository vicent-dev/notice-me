package app

import (
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
)

func (s *server) consumeNotificationHandler() func([]byte) {
	repo := repository.GetRepository[notification.Notification](s.db)
	return func(body []byte) {
		notification.ConsumeNotification(repo, s.ws, body)
	}
}
