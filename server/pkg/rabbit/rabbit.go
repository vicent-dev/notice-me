package rabbit

import (
	"context"
	"github.com/en-vee/alog"
	amqp "github.com/rabbitmq/amqp091-go"
	"notice-me-server/pkg/config"
	"sync"
	"time"
)

type Rabbit struct {
	conn         *amqp.Connection
	QueuesConfig []config.QueueConfig
}

func NewRabbit(conn *amqp.Connection, queuesConfig []config.QueueConfig) *Rabbit {
	return &Rabbit{
		conn:         conn,
		QueuesConfig: queuesConfig,
	}
}

func (r *Rabbit) declareQueues() {
	ch, _ := r.conn.Channel()

	defer ch.Close()

	for _, queue := range r.QueuesConfig {

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
	wg.Add(len(r.QueuesConfig))

	for _, queue := range r.QueuesConfig {
		alog.Info("Consuming from queue " + queue.Name)
		go r.Consume(queue, callbacks)
	}

	wg.Wait()
}

func (r *Rabbit) Consume(queue config.QueueConfig, callbacks map[string]func(body []byte)) {
	ch, _ := r.conn.Channel()

	q, _ := ch.QueueDeclare(
		queue.Name,
		queue.Durable,
		queue.AutoDelete,
		queue.Exclusive,
		queue.NoWait,
		nil,
	)

	msgs, _ := ch.Consume(
		q.Name,
		"",
		true,
		queue.Exclusive,
		false,
		queue.NoWait,
		nil,
	)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			alog.Info("Message received [" + d.RoutingKey + "] " + string(d.Body))

			callback, ok := callbacks[d.RoutingKey]

			if ok {
				callback(d.Body)
			}
		}
	}()

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
