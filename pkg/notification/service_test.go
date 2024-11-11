package notification

import (
	"encoding/json"
	"notice-me-server/pkg/rabbit/mock"
	repo_mock "notice-me-server/pkg/repository/mock"
	"testing"
)

func TestPublishCreateNotification(t *testing.T) {
	rm := mock.NewRabbitMock()

	n, err := PublishCreateNotification(
		&NotificationPostDto{
			Body:          "foobar",
			ClientId:      "*",
			ClientGroupId: "*",
		},
		rm,
	)

	if err != nil {
		t.Fatalf("Fail publish create notification method: %s", err.Error())
	}

	if n.ID.String() == "" {
		t.Fatalf("Fail publish create notification method. Notification ID empty")
	}

	if len(rm.(*mock.Rabbit).ProducedMessages) == 0 {
		t.Fatalf("Fail publish create notification method. Any message produced")
	}
}

func TestPublishNotifyNotification(t *testing.T) {
	rm := mock.NewRabbitMock()

	err := PublishNotifyNotification(
		"fake_id",
		rm,
	)

	if err != nil {
		t.Fatalf("Fail publish notify notification method: %s", err.Error())
	}

	if len(rm.(*mock.Rabbit).ProducedMessages) == 0 {
		t.Fatalf("Fail publish notify notification method. Any message produced")
	}
}

func TestGetNotification(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()

	repo.CreateBulk([]Notification{
		{},
		{},
		{},
	})

	n, err := GetNotification(
		"1",
		repo,
	)

	if err != nil || n == nil {
		t.Fatalf("Fail Get Notification method: %s", err.Error())
	}
}

func TestDeleteNotification(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()

	err := DeleteNotification(
		"1",
		repo,
	)

	if err == nil {
		t.Fatalf("Fail Delete Notification method. It allowed to delete invalid id")
	}

	repo.CreateBulk([]Notification{
		{},
		{},
		{},
	})

	err = DeleteNotification(
		"1",
		repo,
	)

	if err != nil {
		t.Fatalf("Fail Delete Notification method: %s", err.Error())
	}
}

func TestCreateNotification(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()
	rm := mock.NewRabbitMock()

	fn := NewNotification(
		"foo bar",
		"*",
		"*",
		false,
	)

	body, err := json.Marshal(fn)

	if err != nil {
		t.Fatalf("Can not marshal notification: %s", err.Error())
	}

	err = CreateNotification(
		repo,
		rm,
		body,
	)

	if err != nil {
		t.Fatalf("Fail create notification: %s", err.Error())
	}

	nPersisted, err := repo.Find("0")

	if err != nil {
		t.Fatalf("Fail create notification. Notification not persisted: %s", err.Error())
	}

	if nPersisted.ID.String() != fn.ID.String() {
		t.Fatalf("Fail create notification. Notification not persisted. Wrong UUID")
	}

	if len(rm.(*mock.Rabbit).ProducedMessages) != 0 {
		t.Fatalf("Fail create notification method instant false. Messages produced")
	}
}

func TestCreateNotificationInstant(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()
	rm := mock.NewRabbitMock()

	fn := NewNotification(
		"foo bar",
		"*",
		"*",
		true,
	)

	body, err := json.Marshal(fn)

	if err != nil {
		t.Fatalf("Can not marshal notification: %s", err.Error())
	}

	err = CreateNotification(
		repo,
		rm,
		body,
	)

	if err != nil {
		t.Fatalf("Fail create notification: %s", err.Error())
	}

	nPersisted, err := repo.Find("0")

	if err != nil {
		t.Fatalf("Fail create notification. Notification not persisted: %s", err.Error())
	}

	if nPersisted.ID.String() != fn.ID.String() {
		t.Fatalf("Fail create notification. Notification not persisted. Wrong UUID")
	}

	if len(rm.(*mock.Rabbit).ProducedMessages) == 0 {
		t.Fatalf("Fail create notification method instant false. No notify message produced.")
	}
}
