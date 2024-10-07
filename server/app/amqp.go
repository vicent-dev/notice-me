package app

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
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
	consumers["notification.notify"] = s.consumeNotificationNotifyHandler()
	consumers["notification.create"] = s.consumeNotificationCreateHandler()

	return consumers
}
