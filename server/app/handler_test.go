package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/rabbit/mock"
	repo_mock "notice-me-server/pkg/repository/mock"
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

func initialiseMocks() {
	// reset every test the status of the mocks
	rbb := mock.NewRabbitMock()

	repositories := make(map[string]interface{})
	repositories[notification.RepositoryKey] = repo_mock.NewRepository[notification.Notification]()

	s = &server{
		repositories: repositories,
		rabbit:       rbb,
	}
}
