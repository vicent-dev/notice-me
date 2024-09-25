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