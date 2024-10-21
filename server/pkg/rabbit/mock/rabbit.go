package mock

import (
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
)

type Rabbit struct {
	ProducedMessages [][]byte
}

func NewRabbitMock() rabbit.RabbitInterface {
	return &Rabbit{}
}

func (r *Rabbit) Close() error {
	return nil
}

func (r *Rabbit) GetQueuesConfig() []config.QueueConfig {
	return []config.QueueConfig{}
}

func (r *Rabbit) RunConsumers(callbacks map[string]func(body []byte)) {
}

func (r *Rabbit) Consume(queue config.QueueConfig, callbacks map[string]func(body []byte), consumerKey string) {
}

func (r *Rabbit) Produce(queue config.QueueConfig, msg []byte) error {
	r.ProducedMessages = append(r.ProducedMessages, msg)
	return nil
}
