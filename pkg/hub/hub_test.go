package hub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestStopHaltsRun(t *testing.T) {
	h := NewHub()

	done := make(chan struct{})
	go func() {
		h.Run()
		close(done)
	}()

	h.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("hub did not stop after Stop")
	}
}

func TestStopClosesClients(t *testing.T) {
	h := NewHub()
	go h.Run()

	registered := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		h.RegisterClient(NewClient("id", "group", h, conn, make(chan []byte, 1)))
		close(registered)
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	select {
	case <-registered:
	case <-time.After(time.Second):
		t.Fatal("client was not registered")
	}

	h.Stop()

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected client connection to be closed after Stop")
	}
}
