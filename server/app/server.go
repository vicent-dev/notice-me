package app

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"net/http"
	"notice-me-server/app/config"
	"notice-me-server/app/websocket"
)

type server struct {
	r    *mux.Router
	ws   *websocket.Hub
	amqp *amqp.Connection
	db   *gorm.DB
	c    *config.Config
}

func NewServer() *server {
	s := server{
		c:  config.LoadConfig(),
		ws: websocket.NewHub(),
		r:  mux.NewRouter(),
	}

	s.connectDb()
	s.connectAmqp()

	s.routes()

	return &s
}

func (s *server) Run() error {

	go func(websocket *websocket.Hub) {
		websocket.Run()
	}(s.ws)

	return http.ListenAndServe(":"+s.c.Server.Port, handlers.RecoveryHandler()(s.r))
}

func (s *server) writeResponse(w http.ResponseWriter, response map[string]interface{}) {
	w.WriteHeader(http.StatusOK)

	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}

func (s *server) writeErrorResponse(w http.ResponseWriter, response map[string]interface{}, errorCode int) {
	w.WriteHeader(errorCode)
	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}
