package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"time"
)

func connectRabbit() *amqp.Connection {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBIT_USER"),
		os.Getenv("RABBIT_PASSWORD"),
		os.Getenv("RABBIT_HOST"),
		os.Getenv("RABBIT_PORT"),
	))

	if err != nil {
		panic(err)
	}

	return conn
}

func publishNotificationNotify(rabbit *amqp.Connection, n *Notification) {
	ch, _ := rabbit.Channel()

	// When you create the queue in your business implementation be sure that the queue config
	// matches the one declared in Notice-me config file
	q, _ := ch.QueueDeclare(
		"notification.notify",
		true,
		false,
		false,
		false,
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nJson, _ := json.Marshal(n)

	ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        nJson,
		})
}
