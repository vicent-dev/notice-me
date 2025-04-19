package notification

import (
	"encoding/json"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/repository"
	"time"

	"github.com/en-vee/alog"
)

func PublishCreateNotification(
	notificationPostDto *NotificationPostDto,
	rabbit rabbit.RabbitInterface,
) (*Notification, error) {

	n := NewNotification(
		notificationPostDto.Body,
		notificationPostDto.ClientId,
		notificationPostDto.ClientGroupId,
		notificationPostDto.Instant,
		notificationPostDto.OriginClientId,
	)

	var queueConfigCreate config.QueueConfig

	for _, qc := range rabbit.GetQueuesConfig() {
		if qc.Name == "notification.create" {
			queueConfigCreate = qc
		}
	}

	// we publish into rabbitmq the dto because of the api key value
	nBody, err := json.Marshal(notificationPostDto)

	if err != nil {
		return nil, err
	}

	err = rabbit.Produce(queueConfigCreate, nBody)

	if err != nil {
		return nil, err
	}

	return n, nil
}

func PublishNotifyNotification(
	id string,
	rabbit rabbit.RabbitInterface,
) error {

	var queueConfigCreate config.QueueConfig

	for _, qc := range rabbit.GetQueuesConfig() {
		if qc.Name == "notification.notify" {
			queueConfigCreate = qc
		}
	}

	n := &NotificationNotifyDto{id}
	nBody, err := json.Marshal(n)

	if err != nil {
		return err
	}

	err = rabbit.Produce(queueConfigCreate, nBody)

	if err != nil {
		return err
	}

	return nil
}

func GetNotifications(
	repo repository.Repository[Notification],
	pageSize, page int,
) (*repository.Pagination, error) {
	return repo.FindPaginated(pageSize, page)
}

func GetNotification(
	id string,
	repo repository.Repository[Notification],
) (*Notification, error) {
	n, err := repo.Find(id)

	if err != nil {
		return nil, err
	}

	return n, nil
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

func CreateNotification(repo repository.Repository[Notification], apiKeyRepo repository.Repository[auth.ApiKey], rabbit rabbit.RabbitInterface, ws hub.HubInterface, body []byte) error {
	notificationPostDto := &NotificationPostDto{}

	err := json.Unmarshal(body, notificationPostDto)
	if err != nil {
		alog.Error("Cannot unmarshal notification.create: " + err.Error())
		return err
	}

	apiKey, err := apiKeyRepo.FindBy(repository.Field{Column: "Value", Value: notificationPostDto.ApiKeyValue})

	if err != nil {
		alog.Error("Cannot find any api key by the id: " + notificationPostDto.ApiKeyValue)
		return err
	}

	n := NewNotification(
		notificationPostDto.Body,
		notificationPostDto.ClientId,
		notificationPostDto.ClientGroupId,
		notificationPostDto.Instant,
		notificationPostDto.OriginClientId,
	)

	n.ApiKey = apiKey[0]

	err = repo.Create(n)
	if err != nil {
		alog.Error("Cannot create notification: " + err.Error())
		ws.NotifyError(n.OriginClientId, []byte("Cannot create notification: "+err.Error()))
		return err
	}

	if n.Instant {
		err = PublishNotifyNotification(n.ID.String(), rabbit)

		if err != nil {
			alog.Error("Cannot publish notify notification: " + err.Error())
			ws.NotifyError(n.OriginClientId, []byte("Cannot publish notify notification: "+err.Error()))
			return err
		}
	}

	return nil
}

func NotifyNotification(repo repository.Repository[Notification], ws hub.HubInterface, body []byte) {
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

	ws.Notify(n.ClientId, n.ClientGroupId, []byte(n.FormatHTML()))
}
