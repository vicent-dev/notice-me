package app

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/rabbit/mock"
	repo_mock "notice-me-server/pkg/repository/mock"
	"strings"
	"testing"
)

var (
	s *server
)

func TestCreateNotificationHandlerSuccess(t *testing.T) {
	initialiseMocks()

	body := `{
		"body": "foo bar",
		"clientId": "*",
		"clientGroupId": "*"
	}`

	req, err := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer([]byte(body)))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.createNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Fail create notification handler status: %v", status)
	}

	if len(s.rabbit.(*mock.Rabbit).ProducedMessages) != 1 {
		t.Errorf("Fail create notification handler produced messages")
	}
}

func TestCreateNotificationHandlerFail(t *testing.T) {
	initialiseMocks()

	body := `{
		"body": "",
		"clientId": "",
		"clientGroupId": ""
	}`

	req, err := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer([]byte(body)))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.createNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status == http.StatusOK {
		t.Errorf("Fail create notification handler status: %v", status)
	}

	if len(s.rabbit.(*mock.Rabbit).ProducedMessages) > 0 {
		t.Errorf("Fail create notification handler failed. It produced messages")
	}
}

func TestGetNotificationsHandlerSuccess(t *testing.T) {
	initialiseMocks()

	req, err := http.NewRequest(http.MethodGet, "/notifications?pageSize=5&page=1", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.getNotificationsHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Fail get notifications handler status: %v", status)
	}
}

func TestGetNotificationHandlerSuccess(t *testing.T) {
	initialiseMocks()

	req, err := http.NewRequest(http.MethodGet, "/notifications", nil)

	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.getNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Fail get notification handler status: %v", status)
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("body: %v", string(b))
	}
}

func TestDeleteNotificationsHandlerSuccess(t *testing.T) {
	initialiseMocks()

	req, err := http.NewRequest(http.MethodDelete, "/notifications", nil)

	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.deleteNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Fail delete notifications handler status: %v", status)
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("body: %v", string(b))
	}
}

func TestNotifyNotificationsHandlerSuccess(t *testing.T) {
	initialiseMocks()

	req, err := http.NewRequest(http.MethodGet, "/notifications/notify", nil)

	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.notifyNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Fail notify notifications handler status: %v", status)
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("body: %v", string(b))
	}

	if len(s.rabbit.(*mock.Rabbit).ProducedMessages) == 0 {
		t.Errorf("Fail notify notification handler failed. It did not produce messages")
	}
}

func TestDeleteNotificationsHandlerFail(t *testing.T) {
	initialiseMocks()

	req, err := http.NewRequest(http.MethodDelete, "/notifications", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.deleteNotificationHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status == http.StatusOK {
		t.Errorf("Fail get notifications fail handler status: %v", status)
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("body: %v", string(b))
	}
}

func TestWSHandlerSuccess(t *testing.T) {
	initialiseMocks()

	server := httptest.NewServer(http.HandlerFunc(s.wsHandler()))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	client, rr, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not connect to WebSocket server: %v", err)
	}

	defer client.Close()

	if status := rr.StatusCode; status != http.StatusSwitchingProtocols {
		t.Errorf("Fail connect ws fail handler status: %v", status)
		b, _ := io.ReadAll(rr.Body)
		t.Errorf("body: %v", string(b))
	}
}

func initialiseMocks() {
	// reset every test the status of the mocks
	rbb := mock.NewRabbitMock()

	repositories := make(map[string]interface{})
	notificationsRepo := repo_mock.NewRepository[notification.Notification]()

	notificationsRepo.CreateBulk([]notification.Notification{
		{},
		{},
		{},
	})

	repositories[notification.RepositoryKey] = notificationsRepo

	c := &config.Config{}

	c.Server.Cors = []string{"*"}
	s = &server{
		repositories: repositories,
		rabbit:       rbb,
		ws:           hub.NewHub(),
		c:            c,
	}
}
