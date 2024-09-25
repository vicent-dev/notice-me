package notification

import (
	"encoding/json"
	"notice-me-server/app/config"
	"notice-me-server/app/rabbit"
)

func CreateNotification(r *rabbit.Rabbit) (*Notification, error) {
	n := &Notification{
		text: "test",
	}

	nJson, _ := json.Marshal(n)

	r.Produce(config.QueueConfig{
		Name:       "notification.create",
		Exchange:   "",
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}, nJson)

	return n, nil
}
