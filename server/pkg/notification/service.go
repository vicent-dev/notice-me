package notification

import (
	"encoding/json"
	"github.com/en-vee/alog"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/repository"
	"notice-me-server/pkg/websocket"
	"time"
)

func PublishCreateNotification(
	notificationPostDto *NotificationPostDto,
	rabbit *rabbit.Rabbit,
) (*Notification, error) {

	n := NewNotification(
		notificationPostDto.Body,
		notificationPostDto.ClientId,
		notificationPostDto.ClientGroupId,
	)

	var queueConfigCreate config.QueueConfig

	for _, qc := range rabbit.QueuesConfig {
		if qc.Name == "notification.create" {
			queueConfigCreate = qc
		}
	}

	nBody, err := json.Marshal(n)

	if err != nil {
		return nil, err
	}

	err = rabbit.Produce(queueConfigCreate, nBody)

	if err != nil {
		return nil, err
	}

	return n, nil
}

func GetNotifications(
	repo repository.Repository[Notification],
	pageSize, page int,
) (*repository.Pagination, error) {
	return repo.FindPaginated(pageSize, page)
}

func DeleteNotification(
	id string,
	repo repository.Repository[Notification],
) error {

	n, err := repo.Find(id)

	if err != nil {
		return err
	}

	err = repo.Delete(n)

	if err != nil {
		return err
	}

	return nil
}

func CreateNotification(repo repository.Repository[Notification], body []byte) {
	n := &Notification{}

	err := json.Unmarshal(body, n)
	if err != nil {
		alog.Error("Cannot unmarshal notification.create: " + err.Error())
		return
	}

	err = repo.Create(n)
	if err != nil {
		alog.Error("Cannot create notification: " + err.Error())
		return
	}
}

func NotifyNotification(repo repository.Repository[Notification], ws *websocket.Hub, body []byte) {
	//update notification
	queueNotification := &Notification{}

	err := json.Unmarshal(body, queueNotification)

	if err != nil {
		alog.Error("Cannot unmarshal notification.notify: " + err.Error())
		return
	}

	n, err := repo.Find(queueNotification.ID.String())

	if err != nil {
		alog.Error("Error consuming message " + string(body))
		return
	}

	err = repo.Update(n, repository.Field{Column: "NotifiedAt", Value: time.Now()})
	if err != nil {
		alog.Error("Cannot update notification: " + err.Error())
		return
	}

	// broadcast to all clients
	if n.ClientId == websocket.AllClientId || n.ClientGroupId == websocket.AllClientGroupId {
		ws.Broadcast <- []byte(n.FormatHTML())
		return
	}

	clients := ws.GetClientsToNotify(n.ClientId, n.ClientGroupId)

	for _, client := range clients {
		client.Send <- []byte(n.FormatHTML())
	}
}
