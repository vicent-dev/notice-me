package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/rabbit"
	"time"

	"github.com/en-vee/alog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type server struct {
	r            *mux.Router
	ws           hub.HubInterface
	rabbit       rabbit.RabbitInterface
	repositories map[string]interface{}
	db           *gorm.DB
	c            *config.Config

	httpServer http.Server
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

	s.httpServer = newHTTPServer(&s)

	return &s
}

func newHTTPServer(s *server) http.Server {
	headersOk := handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	originsOk := handlers.AllowedOrigins(s.c.Server.Cors)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	handler := handlers.CORS(headersOk, originsOk, methodsOk)(handlers.RecoveryHandler()(s.r))

	return http.Server{
		Addr:              ":" + s.c.Server.Port,
		ErrorLog:          newServerErrorLog(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func (s *server) Run(ctx context.Context) error {

	go func(websocket hub.HubInterface) {
		websocket.Run()
	}(s.ws)

	s.startConsumers()

	errCh := make(chan error, 1)

	go func() {
		var err error

		if s.c.Server.Env == "production" {
			err = s.httpServer.ListenAndServeTLS(s.c.Server.TlsCert, s.c.Server.TlsKey)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()
	alog.Info("shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownErr := s.httpServer.Shutdown(shutdownCtx)

	s.ws.Stop()
	s.rabbit.Stop()

	if err := s.rabbit.Close(); err != nil {
		alog.Error("Error closing amqp connection: " + err.Error())
	}

	if s.db != nil {
		if dbInstance, err := s.db.DB(); err == nil {
			if err := dbInstance.Close(); err != nil {
				alog.Error("Error closing sql connection: " + err.Error())
			}
		}
	}

	if shutdownErr != nil {
		return fmt.Errorf("http server shutdown: %w", shutdownErr)
	}

	return <-errCh
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
