package rabbit

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"notice-me-server/pkg/config"
	"sync"
	"time"
)

type Rabbit struct {
	conn         *amqp.Connection
	queuesConfig []config.QueueConfig
}

func NewRabbit(conn *amqp.Connection, queuesConfig []config.QueueConfig) *Rabbit {
	return &Rabbit{
		conn:         conn,
		queuesConfig: queuesConfig,
	}
}

func (r *Rabbit) declareQueues() {
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

func (r *Rabbit) RunConsumers(callbacks map[string]func(body []byte)) {
	wg := sync.WaitGroup{}
	wg.Add(len(r.queuesConfig))

	for _, queue := range r.queuesConfig {
		go r.Consume(queue, callbacks)
	}

	wg.Wait()
}

func (r *Rabbit) Consume(queue config.QueueConfig, callbacks map[string]func(body []byte)) {
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
			callback, ok := callbacks[d.RoutingKey]

			if ok {
				callback(d.Body)
			}
		}
	}()

	log.Printf(fmt.Sprintf(" [*] Waiting for messages from queue %s. To exit press CTRL+C", queue.Name))

	<-forever
}

func (r *Rabbit) Produce(queue config.QueueConfig, msg []byte) error {
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
