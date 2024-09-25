package app

import (
	"log"
	"net/http"
	"notice-me-server/app/rabbit"
	"notice-me-server/app/websocket"
	"notice-me-server/pkg/notification"
)

func (s *server) routes() {
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(jsonMiddleware)

	s.r.Use(loggingMiddleware)

	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	//notify POST
	apiRouter.HandleFunc("/notification", s.createNotificationHandler()).Methods("POST")
}

// handlers @todo move if needed
func (s *server) createNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	rab := rabbit.NewRabbit(s.amqp, s.c.Rabbit.Queues)
	return func(w http.ResponseWriter, r *http.Request) {
		notification.CreateNotification(rab)
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
