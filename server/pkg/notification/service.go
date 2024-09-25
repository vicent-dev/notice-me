package notification

import (
	"encoding/json"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/websocket"
)

func CreateNotification(r *rabbit.Rabbit) (*Notification, error) {
	n := &Notification{
		Body: "test body",
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

func ConsumeNotification(ws *websocket.Hub, body []byte) {
	// broadcast to all clients
	ws.Broadcast <- body
}
