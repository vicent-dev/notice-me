package app

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"time"
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

func TestDocsHandlerServesHTML(t *testing.T) {
	initialiseMocks()

	rr := httptest.NewRecorder()
	http.HandlerFunc(s.docsHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/docs", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("docs handler returned an empty body")
	}
}

func TestCreateNotificationHandlerMalformedJSON(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString("{not-json"))
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.createNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if len(s.rabbit.(*mock.Rabbit).ProducedMessages) != 0 {
		t.Fatal("malformed request must not publish a message")
	}
}

func TestCreateNotificationHandlerRabbitError(t *testing.T) {
	initialiseMocks()
	s.rabbit.(*mock.Rabbit).ProduceErr = errors.New("broker down")

	body := `{"body":"x","clientId":"*","clientGroupId":"*"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.createNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetNotificationsHandlerInvalidPageSize(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodGet, "/notifications?pageSize=abc", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.getNotificationsHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetNotificationsHandlerInvalidPage(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodGet, "/notifications?page=abc", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.getNotificationsHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetNotificationsHandlerRepoError(t *testing.T) {
	initialiseMocks()

	repo := repo_mock.NewErrorRepository[notification.Notification]()
	repo.FindPaginatedErr = errors.New("db down")
	s.repositories[notification.RepositoryKey] = repo

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.getNotificationsHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetNotificationHandlerNotFound(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodGet, "/notifications/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.getNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteNotificationHandlerNotFound(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodDelete, "/notifications/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.deleteNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotifyNotificationHandlerNotFound(t *testing.T) {
	initialiseMocks()

	req := httptest.NewRequest(http.MethodGet, "/notifications/notify/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.notifyNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotifyNotificationHandlerAlreadyNotified(t *testing.T) {
	initialiseMocks()

	notifiedAt := time.Now()
	n := notification.NewNotification("x", "c", "g", false, "")
	n.NotifiedAt = &notifiedAt

	repo := repo_mock.NewRepository[notification.Notification]()
	repo.Create(n)
	s.repositories[notification.RepositoryKey] = repo

	req := httptest.NewRequest(http.MethodGet, "/notifications/notify/0", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "0"})
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.notifyNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotifyNotificationHandlerRabbitError(t *testing.T) {
	initialiseMocks()
	s.rabbit.(*mock.Rabbit).ProduceErr = errors.New("publish failed")

	repo := repo_mock.NewRepository[notification.Notification]()
	repo.Create(notification.NewNotification("x", "c", "g", false, ""))
	s.repositories[notification.RepositoryKey] = repo

	req := httptest.NewRequest(http.MethodGet, "/notifications/notify/0", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "0"})
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.notifyNotificationHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestWriteResponseNilReturnsCreated(t *testing.T) {
	initialiseMocks()

	rr := httptest.NewRecorder()
	s.writeResponse(rr, nil)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
}

func TestWriteErrorResponseShape(t *testing.T) {
	initialiseMocks()

	rr := httptest.NewRecorder()
	s.writeErrorResponse(rr, errors.New("nope"), http.StatusBadRequest)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if payload["error"] != "nope" {
		t.Fatalf("error = %q, want %q", payload["error"], "nope")
	}
}

func TestWSHandlerRejectsDisallowedOrigin(t *testing.T) {
	initialiseMocks()
	s.c.Server.Cors = []string{"allowed.example.com"}

	server := httptest.NewServer(http.HandlerFunc(s.wsHandler()))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	header := http.Header{}
	header.Set("Origin", "http://evil.example.com")

	client, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err == nil {
		if client != nil {
			client.Close()
		}
		t.Fatal("expected handshake to fail for a disallowed origin")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %v", resp)
	}
}
