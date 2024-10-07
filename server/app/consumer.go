package app

import (
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
)

func (s *server) consumeNotificationNotifyHandler() func([]byte) {
	repo := repository.GetRepository[notification.Notification](s.db)
	return func(body []byte) {
		notification.NotifyNotification(repo, s.ws, body)
	}
}

func (s *server) consumeNotificationCreateHandler() func([]byte) {
	repo := repository.GetRepository[notification.Notification](s.db)
	return func(body []byte) {
		notification.CreateNotification(repo, body)
	}
}
