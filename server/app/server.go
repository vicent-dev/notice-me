package app

import (
	"encoding/json"
	"net/http"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/websocket"

	"github.com/en-vee/alog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type server struct {
	r            *mux.Router
	ws           websocket.HubInterface
	rabbit       rabbit.RabbitInterface
	repositories map[string]interface{}
	db           *gorm.DB
	c            *config.Config
}

func NewServer() *server {
	s := server{
		c:  config.LoadConfig(),
		ws: websocket.NewHub(),
		r:  mux.NewRouter(),
	}

	s.initialiseRepositories()
	s.initialiseRabbit()

	s.routes()

	return &s
}

func (s *server) Run() error {

	defer func() {
		err := s.rabbit.Close()
		if err != nil {
			alog.Error("Error closing amqp connection: " + err.Error())
		}
	}()

	defer func() {
		dbInstance, _ := s.db.DB()
		err := dbInstance.Close()
		if err != nil {
			alog.Error("Error closing sql connection: " + err.Error())
		}
	}()

	go func(websocket websocket.HubInterface) {
		websocket.Run()
	}(s.ws)

	go func(r rabbit.RabbitInterface, consumers map[string]func([]byte)) {
		r.RunConsumers(consumers)
	}(s.rabbit, s.consumersMap())

	headersOk := handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	originsOk := handlers.AllowedOrigins(s.c.Server.Cors)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

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
