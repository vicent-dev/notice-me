package app

import (
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
		notification.CreateNotification(repo, rab)
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
