package app

import (
	"encoding/json"
	"errors"
	"testing"

	hub_mock "notice-me-server/pkg/hub/mock"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	repo_mock "notice-me-server/pkg/repository/mock"
)

// uuidRepo is a fake repository.Repository[Notification] keyed by UUID. The
// shared counter-keyed mock cannot serve NotifyNotification, which looks up
// rows by the notification's UUID.
type uuidRepo struct {
	entities map[string]*notification.Notification
}

func newUUIDRepo() *uuidRepo {
	return &uuidRepo{entities: make(map[string]*notification.Notification)}
}

func (r *uuidRepo) Find(id string) (*notification.Notification, error) {
	if n, ok := r.entities[id]; ok {
		return n, nil
	}
	return nil, errors.New("entity not found")
}

func (r *uuidRepo) Create(t *notification.Notification) error {
	r.entities[t.ID.String()] = t
	return nil
}

func (r *uuidRepo) Update(t *notification.Notification, fs ...repository.Field) error {
	r.entities[t.ID.String()] = t
	return nil
}

func (r *uuidRepo) Delete(t *notification.Notification) error {
	delete(r.entities, t.ID.String())
	return nil
}

func (r *uuidRepo) FindPaginated(pageSize, page int) (*repository.Pagination, error) {
	return &repository.Pagination{Limit: pageSize, Page: page}, nil
}

func (r *uuidRepo) FindWithRelations(id string) (*notification.Notification, error) {
	return r.Find(id)
}
func (r *uuidRepo) FindBy(fs ...repository.Field) ([]*notification.Notification, error) {
	return nil, nil
}
func (r *uuidRepo) FindByWithRelations(fs ...repository.Field) ([]*notification.Notification, error) {
	return nil, nil
}
func (r *uuidRepo) FindFirstBy(fs ...repository.Field) (*notification.Notification, error) {
	return nil, errors.New("entity not found")
}
func (r *uuidRepo) CreateBulk(ts []notification.Notification) error {
	for i := range ts {
		r.Create(&ts[i])
	}
	return nil
}

func TestConsumeNotificationCreateHandlerPersists(t *testing.T) {
	initialiseMocks()

	repo := repo_mock.NewRepository[notification.Notification]()
	s.repositories[notification.RepositoryKey] = repo

	n := notification.NewNotification("body", "c", "g", false, "origin")
	body, _ := json.Marshal(n)

	s.consumeNotificationCreateHandler()(body)

	got, err := repo.Find("0")
	if err != nil {
		t.Fatalf("notification not persisted: %v", err)
	}
	if got.Body != "body" {
		t.Fatalf("Body = %q, want %q", got.Body, "body")
	}
}

func TestConsumeNotificationNotifyHandlerNotifiesHub(t *testing.T) {
	initialiseMocks()

	repo := newUUIDRepo()
	s.repositories[notification.RepositoryKey] = repo

	ws := hub_mock.NewHubMock().(*hub_mock.Hub)
	s.ws = ws

	n := notification.NewNotification("payload", "c1", "g1", false, "")
	repo.Create(n)
	body, _ := json.Marshal(n)

	s.consumeNotificationNotifyHandler()(body)

	if len(ws.Notified) != 1 {
		t.Fatalf("expected 1 Notify call, got %d", len(ws.Notified))
	}
	if ws.Notified[0].ClientId != "c1" || ws.Notified[0].ClientGroupId != "g1" {
		t.Fatalf("Notify targets = (%q,%q), want (c1,g1)",
			ws.Notified[0].ClientId, ws.Notified[0].ClientGroupId)
	}
	if string(ws.Notified[0].Body) != "payload" {
		t.Fatalf("Notify body = %q, want %q", ws.Notified[0].Body, "payload")
	}
}
