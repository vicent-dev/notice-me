package app

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"net/http"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/websocket"
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

	go func(amqp *amqp.Connection, queues []config.QueueConfig, consumers map[string]func([]byte)) {
		r := rabbit.NewRabbit(amqp, queues)
		r.RunConsumers(consumers)
	}(s.amqp, s.c.Rabbit.Queues, s.consumersMap())

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins(s.c.Server.Cors)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	handler := handlers.CORS(headersOk, originsOk, methodsOk)(handlers.RecoveryHandler()(s.r))

	if s.c.Server.Env == "production" {

		return http.ListenAndServeTLS(":"+s.c.Server.Port, s.c.Server.TlsCert, s.c.Server.TlsKey, handler)
	} else {
		return http.ListenAndServe(":"+s.c.Server.Port, handler)
	}
}

func (s *server) writeResponse(w http.ResponseWriter, response interface{}) {
	if response == nil {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}

func (s *server) writeErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := make(map[string]interface{})

	response["error"] = err.Error()
	w.WriteHeader(errorCode)
	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}
