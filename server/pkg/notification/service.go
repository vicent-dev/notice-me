package notification

import (
	"encoding/json"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/repository"
	"notice-me-server/pkg/websocket"
	"time"
)

func CreateNotification(
	notificationPostDto *NotificationPostDto,
	repo repository.Repository[Notification],
	r *rabbit.Rabbit,
) (*Notification, error) {
	n := &Notification{
		Body: notificationPostDto.Body,
	}

	repo.Create(n)

	nJson, _ := json.Marshal(n)

	var queueConfig config.QueueConfig

	for _, qc := range r.QueuesConfig {
		if qc.Name == "notification.create" {
			queueConfig = qc
		}
	}

	r.Produce(queueConfig, nJson)

	return n, nil
}

func ConsumeNotification(repo repository.Repository[Notification], ws *websocket.Hub, body []byte) {

	//update notification
	n := &Notification{}

	json.Unmarshal(body, n)

	repo.Find(n.ID)
	repo.Update(n, repository.Field{Column: "NotifiedAt", Value: time.Now()})

	// broadcast to all clients
	ws.Broadcast <- []byte(n.FormatHTML())
}
