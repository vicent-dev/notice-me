package app

import (
	"encoding/json"
	"github.com/en-vee/alog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/rabbit"
)

type server struct {
	r            *mux.Router
	ws           hub.HubInterface
	rabbit       rabbit.RabbitInterface
	repositories map[string]interface{}
	db           *gorm.DB
	c            *config.Config
}

func NewServer() *server {
	s := server{
		c:  config.LoadConfig(),
		ws: hub.NewHub(),
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

	go func(websocket hub.HubInterface) {
		websocket.Run()
	}(s.ws)

	s.startConsumers()

	headersOk := handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	originsOk := handlers.AllowedOrigins(s.c.Server.Cors)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	handler := handlers.CORS(headersOk, originsOk, methodsOk)(handlers.RecoveryHandler()(s.r))

	log := newServerErrorLog()

	server := &http.Server{
		Addr:     ":" + s.c.Server.Port,
		ErrorLog: log,
		Handler:  handler,
	}

	if s.c.Server.Env == "production" {
		return server.ListenAndServeTLS(s.c.Server.TlsCert, s.c.Server.TlsKey)
	} else {
		return server.ListenAndServe()
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
