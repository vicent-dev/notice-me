package notification

import (
	"encoding/json"
	"github.com/en-vee/alog"
	"notice-me-server/pkg/repository"
	"notice-me-server/pkg/websocket"
	"strconv"
	"time"
)

func CreateNotification(
	notificationPostDto *NotificationPostDto,
	repo repository.Repository[Notification],
) (*Notification, error) {
	n := &Notification{
		Body:          notificationPostDto.Body,
		ClientId:      notificationPostDto.ClientId,
		ClientGroupId: notificationPostDto.ClientGroupId,
	}

	repo.Create(n)

	return n, nil
}

func GetNotifications(
	repo repository.Repository[Notification],
) ([]*Notification, error) {

	return repo.FindBy()
}

func DeleteNotification(
	id string,
	repo repository.Repository[Notification],
) error {
	idInt, err := strconv.Atoi(id)

	if err != nil {
		return err
	}

	n, err := repo.Find(uint(idInt))

	if err != nil {
		return err
	}

	err = repo.Delete(n)

	if err != nil {
		return err
	}

	return nil
}

func ConsumeNotification(repo repository.Repository[Notification], ws *websocket.Hub, body []byte) {
	//update notification
	queueNotification := &Notification{}

	json.Unmarshal(body, queueNotification)

	n, err := repo.Find(queueNotification.ID)

	if err != nil {
		alog.Error("Error consuming message " + string(body))
		return
	}

	repo.Update(n, repository.Field{Column: "NotifiedAt", Value: time.Now()})

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
