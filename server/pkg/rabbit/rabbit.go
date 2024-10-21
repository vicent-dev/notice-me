package rabbit

import (
	"context"
	"notice-me-server/pkg/config"
	"strconv"
	"sync"
	"time"

	"github.com/en-vee/alog"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitInterface interface {
	RunConsumers(callbacks map[string]func(body []byte))
	Consume(queue config.QueueConfig, callbacks map[string]func(body []byte), consumerKey string)
	Produce(queue config.QueueConfig, msg []byte) error
	GetQueuesConfig() []config.QueueConfig
}

type Rabbit struct {
	conn           *amqp.Connection
	consumersCount int
	queuesConfig   []config.QueueConfig
}

func NewRabbit(conn *amqp.Connection, consumersCount int, queuesConfig []config.QueueConfig) RabbitInterface {
	return &Rabbit{
		conn:           conn,
		consumersCount: consumersCount,
		queuesConfig:   queuesConfig,
	}
}

func (r *Rabbit) GetQueuesConfig() []config.QueueConfig {
	return r.queuesConfig
}

func (r *Rabbit) RunConsumers(callbacks map[string]func(body []byte)) {
	wg := sync.WaitGroup{}
	wg.Add(len(r.GetQueuesConfig()) * r.consumersCount)

	for _, queue := range r.queuesConfig {
		consumerKey := queue.Name + "_consumer_group_key"
		for cc := range r.consumersCount {
			alog.Info("Consumer [" + strconv.Itoa(cc) + "] consuming from queue " + queue.Name)
			go r.Consume(queue, callbacks, consumerKey)
		}
	}

	wg.Wait()
}

func (r *Rabbit) Consume(queue config.QueueConfig, callbacks map[string]func(body []byte), consumerKey string) {
	ch, _ := r.conn.Channel()

	defer func() {
		if err := ch.Close(); err != nil {
			alog.Error("Error closing channel: " + err.Error())
		}
	}()

	err := ch.ExchangeDeclare(
		queue.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		alog.Error("Error declaring exchange consumer: " + err.Error())
	}

	q, _ := ch.QueueDeclare(
		queue.Name,
		queue.Durable,
		queue.AutoDelete,
		queue.Exclusive,
		queue.NoWait,
		nil,
	)

	err = ch.QueueBind(q.Name,
		queue.Name,
		queue.Exchange,
		false,
		nil,
	)

	if err != nil {
		alog.Error("Error binding queue consumer: " + err.Error())
		return
	}

	msgs, _ := ch.Consume(
		q.Name,
		consumerKey,
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

	defer func() {
		if err := ch.Close(); err != nil {
			alog.Error("Error closing channel: " + err.Error())
		}
	}()

	err := ch.ExchangeDeclare(
		queue.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

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
		queue.Exchange,
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
}
