package notification

import (
	"encoding/json"
	"errors"
	"notice-me-server/pkg/config"
	hub_mock "notice-me-server/pkg/hub/mock"
	"notice-me-server/pkg/rabbit/mock"
	"notice-me-server/pkg/repository"
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
		"",
	)

	body, err := json.Marshal(fn)

	if err != nil {
		t.Fatalf("Can not marshal notification: %s", err.Error())
	}

	err = CreateNotification(
		repo,
		rm,
		nil,
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
		"",
	)

	body, err := json.Marshal(fn)

	if err != nil {
		t.Fatalf("Can not marshal notification: %s", err.Error())
	}

	err = CreateNotification(
		repo,
		rm,
		nil,
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

// uuidRepo is a fake repository.Repository[Notification] keyed by the entity's
// UUID, mirroring how the real GORM adapter looks up rows (WHERE id = ?). It lets
// domain tests exercise NotifyNotification independently of the database.
type uuidRepo struct {
	entities    map[string]*Notification
	updateErr   error
	updateCalls int
	lastFields  []repository.Field
}

func newUUIDRepo() *uuidRepo {
	return &uuidRepo{entities: make(map[string]*Notification)}
}

func (r *uuidRepo) Find(id string) (*Notification, error) {
	if n, ok := r.entities[id]; ok {
		return n, nil
	}
	return nil, errors.New("entity not found")
}

func (r *uuidRepo) Create(t *Notification) error {
	r.entities[t.ID.String()] = t
	return nil
}

func (r *uuidRepo) Update(t *Notification, fs ...repository.Field) error {
	r.updateCalls++
	r.lastFields = fs
	if r.updateErr != nil {
		return r.updateErr
	}
	r.entities[t.ID.String()] = t
	return nil
}

func (r *uuidRepo) Delete(t *Notification) error {
	delete(r.entities, t.ID.String())
	return nil
}

func (r *uuidRepo) FindPaginated(pageSize, page int) (*repository.Pagination, error) {
	var rows []*Notification
	for _, n := range r.entities {
		rows = append(rows, n)
	}
	return &repository.Pagination{Limit: pageSize, Page: page, Rows: rows, TotalRows: int64(len(rows))}, nil
}

func (r *uuidRepo) FindWithRelations(id string) (*Notification, error)     { return r.Find(id) }
func (r *uuidRepo) FindBy(fs ...repository.Field) ([]*Notification, error) { return nil, nil }
func (r *uuidRepo) FindByWithRelations(fs ...repository.Field) ([]*Notification, error) {
	return nil, nil
}
func (r *uuidRepo) FindFirstBy(fs ...repository.Field) (*Notification, error) {
	return nil, errors.New("entity not found")
}
func (r *uuidRepo) CreateBulk(ts []Notification) error {
	for i := range ts {
		r.Create(&ts[i])
	}
	return nil
}

func TestPublishCreateNotificationProducesPayload(t *testing.T) {
	rm := mock.NewRabbitMock().(*mock.Rabbit)
	rm.QueuesConfig = []config.QueueConfig{{Name: "notification.create"}}

	n, err := PublishCreateNotification(&NotificationPostDto{
		Body:          "foobar",
		ClientId:      "c",
		ClientGroupId: "g",
	}, rm)

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	var produced Notification
	if err := json.Unmarshal(rm.ProducedMessages[0], &produced); err != nil {
		t.Fatalf("produced message is not a notification: %s", err.Error())
	}

	if produced.ID != n.ID {
		t.Fatalf("produced ID %s != returned ID %s", produced.ID, n.ID)
	}
}

func TestPublishCreateNotificationProduceError(t *testing.T) {
	rm := mock.NewRabbitMock().(*mock.Rabbit)
	rm.ProduceErr = errors.New("broker down")

	_, err := PublishCreateNotification(&NotificationPostDto{Body: "b", ClientId: "c", ClientGroupId: "g"}, rm)

	if err == nil {
		t.Fatal("expected error when produce fails, got nil")
	}
}

func TestPublishNotifyNotificationProducesID(t *testing.T) {
	rm := mock.NewRabbitMock().(*mock.Rabbit)
	rm.QueuesConfig = []config.QueueConfig{{Name: "notification.notify"}}

	if err := PublishNotifyNotification("the-id", rm); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	var payload map[string]string
	if err := json.Unmarshal(rm.ProducedMessages[0], &payload); err != nil {
		t.Fatalf("produced message is not valid: %s", err.Error())
	}

	if payload["id"] != "the-id" {
		t.Fatalf("produced id = %q, want %q", payload["id"], "the-id")
	}
}

func TestPublishNotifyNotificationProduceError(t *testing.T) {
	rm := mock.NewRabbitMock().(*mock.Rabbit)
	rm.ProduceErr = errors.New("broker down")

	if err := PublishNotifyNotification("id", rm); err == nil {
		t.Fatal("expected error when produce fails, got nil")
	}
}

func TestGetNotifications(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()
	repo.CreateBulk([]Notification{{}, {}, {}})

	p, err := GetNotifications(repo, 5, 1)
	if err != nil || p == nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}
	if p.TotalRows != 3 {
		t.Fatalf("TotalRows = %d, want 3", p.TotalRows)
	}
}

func TestGetNotificationsError(t *testing.T) {
	repo := repo_mock.NewErrorRepository[Notification]()
	repo.FindPaginatedErr = errors.New("db error")

	if _, err := GetNotifications(repo, 5, 1); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteNotificationFindError(t *testing.T) {
	repo := repo_mock.NewErrorRepository[Notification]()
	repo.FindErr = errors.New("db error")

	if err := DeleteNotification("1", repo); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteNotificationDeleteError(t *testing.T) {
	repo := repo_mock.NewErrorRepository[Notification]()
	repo.Create(&Notification{})
	repo.DeleteErr = errors.New("delete failed")

	if err := DeleteNotification("0", repo); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateNotificationBadJSON(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()
	rm := mock.NewRabbitMock()

	if err := CreateNotification(repo, rm, hub_mock.NewHubMock(), []byte("not-json")); err == nil {
		t.Fatal("expected error on bad json, got nil")
	}
}

func TestCreateNotificationCreateErrorNotifiesOrigin(t *testing.T) {
	repo := repo_mock.NewErrorRepository[Notification]()
	repo.CreateErr = errors.New("insert failed")

	ws := hub_mock.NewHubMock().(*hub_mock.Hub)
	n := NewNotification("foo", "*", "*", false, "origin-1")
	body, _ := json.Marshal(n)

	err := CreateNotification(repo, mock.NewRabbitMock(), ws, body)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(ws.Errors) != 1 {
		t.Fatalf("expected 1 NotifyError call, got %d", len(ws.Errors))
	}
	if ws.Errors[0].OriginClientId != "origin-1" {
		t.Fatalf("NotifyError origin = %q, want %q", ws.Errors[0].OriginClientId, "origin-1")
	}
}

func TestCreateNotificationInstantPublishErrorNotifiesOrigin(t *testing.T) {
	repo := repo_mock.NewRepository[Notification]()
	rm := mock.NewRabbitMock().(*mock.Rabbit)
	rm.ProduceErr = errors.New("publish failed")

	ws := hub_mock.NewHubMock().(*hub_mock.Hub)
	n := NewNotification("foo", "*", "*", true, "origin-2")
	body, _ := json.Marshal(n)

	err := CreateNotification(repo, rm, ws, body)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(ws.Errors) != 1 {
		t.Fatalf("expected 1 NotifyError call, got %d", len(ws.Errors))
	}
	if ws.Errors[0].OriginClientId != "origin-2" {
		t.Fatalf("NotifyError origin = %q, want %q", ws.Errors[0].OriginClientId, "origin-2")
	}
}

func TestNotifyNotificationSuccess(t *testing.T) {
	repo := newUUIDRepo()
	ws := hub_mock.NewHubMock().(*hub_mock.Hub)

	n := NewNotification("hello", "client1", "group1", false, "")
	repo.Create(n)

	NotifyNotification(repo, ws, mustMarshal(t, n))

	if repo.updateCalls != 1 {
		t.Fatalf("updateCalls = %d, want 1", repo.updateCalls)
	}
	if len(ws.Notified) != 1 {
		t.Fatalf("expected 1 Notify call, got %d", len(ws.Notified))
	}
	if ws.Notified[0].ClientId != "client1" || ws.Notified[0].ClientGroupId != "group1" {
		t.Fatalf("Notify targets = (%q,%q), want (client1,group1)",
			ws.Notified[0].ClientId, ws.Notified[0].ClientGroupId)
	}
	if string(ws.Notified[0].Body) != "hello" {
		t.Fatalf("Notify body = %q, want %q", ws.Notified[0].Body, "hello")
	}
}

func TestNotifyNotificationBadJSON(t *testing.T) {
	repo := newUUIDRepo()
	ws := hub_mock.NewHubMock().(*hub_mock.Hub)

	NotifyNotification(repo, ws, []byte("not-json"))

	if len(ws.Notified) != 0 || repo.updateCalls != 0 {
		t.Fatal("bad JSON should not notify or update")
	}
}

func TestNotifyNotificationNotFound(t *testing.T) {
	repo := newUUIDRepo()
	ws := hub_mock.NewHubMock().(*hub_mock.Hub)

	n := NewNotification("hello", "c", "g", false, "")
	NotifyNotification(repo, ws, mustMarshal(t, n))

	if len(ws.Notified) != 0 || repo.updateCalls != 0 {
		t.Fatal("missing notification should not notify or update")
	}
}

func TestNotifyNotificationUpdateErrorSkipsNotify(t *testing.T) {
	repo := newUUIDRepo()
	repo.updateErr = errors.New("db down")
	ws := hub_mock.NewHubMock().(*hub_mock.Hub)

	n := NewNotification("hello", "c", "g", false, "")
	repo.Create(n)

	NotifyNotification(repo, ws, mustMarshal(t, n))

	if repo.updateCalls != 1 {
		t.Fatalf("updateCalls = %d, want 1", repo.updateCalls)
	}
	if len(ws.Notified) != 0 {
		t.Fatal("failed update should not notify clients")
	}
}

func mustMarshal(t *testing.T, n *Notification) []byte {
	t.Helper()
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("cannot marshal notification: %s", err.Error())
	}
	return b
}
