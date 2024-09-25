package app

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

type rabbit struct {
	conn         *amqp.Connection
	queuesConfig map[string]QueueConfig
}

func (s *Server) newRabbit() *rabbit {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		s.c.Rabbit.User,
		s.c.Rabbit.Pwd,
		s.c.Rabbit.Host,
		s.c.Rabbit.Port,
	))

	if err != nil {
		panic(err)
	}

	return &rabbit{
		conn:         conn,
		queuesConfig: s.c.Rabbit.Queues,
	}
}

func (r *rabbit) declareQueues() {
	ch, _ := r.conn.Channel()

	defer ch.Close()

	for _, queue := range r.queuesConfig {

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

func (r *rabbit) runConsummers() {
	wg := sync.WaitGroup{}
	wg.Add(len(r.queuesConfig))

	for _, queue := range r.queuesConfig {
		go r.consume(queue)
	}

	wg.Wait()
}

func (r *rabbit) consume(queue QueueConfig) {
	ch, _ := r.conn.Channel()

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

func (r *rabbit) produce(queue QueueConfig, msg []byte) error {
	ch, _ := r.conn.Channel()

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
