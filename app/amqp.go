package app

import (
	"fmt"
	"github.com/en-vee/alog"
	"notice-me-server/pkg/rabbit"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (s *server) reconnect() {

	go func() {
		for {
			reason, ok := <-s.rabbit.GetConnection().NotifyClose(make(chan *amqp.Error))
			if !ok {
				alog.Info("rabbitmq connection closed")
				break
			}

			alog.Info("rabbitmq connection closed unexpectedly, reason: %v", reason)

			for {

				time.Sleep(time.Duration(1) * time.Second)

				connection, err := s.dialAmqp()

				if err == nil {
					s.rabbit.SetConnection(connection)
					s.startConsumers()
					alog.Info("rabbitmq reconnect success")
					break
				}

				alog.Info("rabbitmq reconnect failed, err: %v", err)
			}

		}
	}()
}

func (s *server) dialAmqp() (*amqp.Connection, error) {
	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		s.c.Rabbit.User,
		s.c.Rabbit.Pwd,
		s.c.Rabbit.Host,
		s.c.Rabbit.Port,
	))
}

func (s *server) initialiseRabbit() {
	conn, err := s.dialAmqp()

	if err != nil {
		panic(err)
	}

	s.rabbit = rabbit.NewRabbit(conn, s.c.Rabbit.ConsumersCount, s.c.Rabbit.Queues)

	s.reconnect()
}

func (s *server) startConsumers() {
	go func(r rabbit.RabbitInterface, consumers map[string]func([]byte)) {
		r.RunConsumers(consumers)
	}(s.rabbit, s.consumersMap())
}

func (s *server) consumersMap() map[string]func([]byte) {
	consumers := make(map[string]func([]byte))
	consumers["notification.notify"] = s.consumeNotificationNotifyHandler()
	consumers["notification.create"] = s.consumeNotificationCreateHandler()

	return consumers
}
