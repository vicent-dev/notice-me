package app

import (
	"encoding/json"
	"gorm.io/gorm"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	r    *mux.Router
	c    *config
	db   *gorm.DB
	amqp *amqp.Connection
}

func NewServer() *server {

	s := server{
		c: loadConfig(),
		r: mux.NewRouter(),
	}

	if s.c.Rabbit.Enabled {
		s.rabbit()
	}
	s.database()
	s.routes()

	return &s
}

func (s *server) Run() error {
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
