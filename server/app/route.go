package app

import (
	"log"
	"net/http"
	"time"
)

func (s *server) routes() {

	s.r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "../client/index.html")
	})

	s.r.Use(loggingMiddleware)

	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")
	s.r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		s.PingAllClients([]byte("test " + time.Now().String()))
	}).Methods("GET")
}

// handlers @todo move if needed
func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		client := &Client{websocketService: ws, conn: conn, send: make(chan []byte, 256)}
		client.websocketService.register <- client

		go client.writePump()
	}
}
