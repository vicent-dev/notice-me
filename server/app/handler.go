package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/repository"
	"notice-me-server/pkg/websocket"
)

func (s *server) createNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	rab := rabbit.NewRabbit(s.amqp, s.c.Rabbit.Queues)
	repo := repository.GetRepository[notification.Notification](s.db)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		notificationPostDto := &notification.NotificationPostDto{}
		if err != json.Unmarshal(body, notificationPostDto) {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		notification.CreateNotification(notificationPostDto, repo, rab)
		s.writeResponse(w, nil)
	}
}

func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws

	return func(w http.ResponseWriter, r *http.Request) {

		websocket.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		conn, err := websocket.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err)
			return
		}
		client := &websocket.Client{WebsocketService: ws, Conn: conn, Send: make(chan []byte, 256)}
		client.WebsocketService.Register <- client

		go client.Write()
	}
}
