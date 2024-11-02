package app

import (
	"fmt"
	"notice-me-server/pkg/rabbit"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (s *server) initialiseRabbit() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		s.c.Rabbit.User,
		s.c.Rabbit.Pwd,
		s.c.Rabbit.Host,
		s.c.Rabbit.Port,
	))

	if err != nil {
		panic(err)
	}

	s.rabbit = rabbit.NewRabbit(conn, s.c.Rabbit.ConsumersCount, s.c.Rabbit.Queues)
}

func (s *server) consumersMap() map[string]func([]byte) {
	consumers := make(map[string]func([]byte))
	consumers["notification.notify"] = s.consumeNotificationNotifyHandler()
	consumers["notification.create"] = s.consumeNotificationCreateHandler()

	return consumers
}
