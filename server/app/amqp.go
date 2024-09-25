package app

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"notice-me-server/pkg/notification"
)

func (s *server) connectAmqp() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		s.c.Rabbit.User,
		s.c.Rabbit.Pwd,
		s.c.Rabbit.Host,
		s.c.Rabbit.Port,
	))

	if err != nil {
		panic(err)
	}

	s.amqp = conn
}

func (s *server) consumersMap() map[string]func([]byte) {
	consumers := make(map[string]func([]byte))
	consumers["notification.create"] = s.consumeNotificationHandler()

	return consumers
}

func (s *server) consumeNotificationHandler() func([]byte) {
	return func(body []byte) {
		notification.ConsumeNotification(s.ws, body)
	}
}
