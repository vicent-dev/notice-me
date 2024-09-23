package app

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

func (s *server) rabbit() {
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

func (s *server) DeclareQueues() {
	ch, _ := s.amqp.Channel()

	defer ch.Close()

	for _, queue := range s.c.Rabbit.Queues {

		_, _ = ch.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			nil,
		)
	}
}

func (s *server) RunConsummers() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.c.Rabbit.Queues))

	for _, queue := range s.c.Rabbit.Queues {
		go s.consume(queue)
	}

	wg.Wait()
}

func (s *server) consume(queue QueueConfig) {
	ch, _ := s.amqp.Channel()

	msgs, _ := ch.Consume(
		queue.Name, // queue
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(fmt.Sprintf(" [*] Waiting for messages from queue %s. To exit press CTRL+C", queue.Name))
	<-forever
}

func (s *server) produce(queue QueueConfig, msg []byte) error {
	ch, _ := s.amqp.Channel()

	q, _ := ch.QueueDeclare(
		queue.Name,
		queue.Durable,
		queue.AutoDelete,
		queue.Exclusive,
		queue.NoWait,
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
}
