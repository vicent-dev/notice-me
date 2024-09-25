package app

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	r      *mux.Router
	ws     *hub
	rabbit *rabbit
	db     *db
	c      *config
}

func NewServer() *Server {
	s := Server{
		c:  loadConfig(),
		ws: newHub(),
		r:  mux.NewRouter(),
	}

	s.db = s.newDb()
	s.rabbit = s.newRabbit()
	s.routes()

	return &s
}

func (s *Server) Run() error {

	go func(websocket *hub) {
		websocket.run()
	}(s.ws)

	return http.ListenAndServe(":"+s.c.Server.Port, handlers.RecoveryHandler()(s.r))
}

func (s *Server) writeResponse(w http.ResponseWriter, response map[string]interface{}) {
	w.WriteHeader(http.StatusOK)

	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}

func (s *Server) writeErrorResponse(w http.ResponseWriter, response map[string]interface{}, errorCode int) {
	w.WriteHeader(errorCode)
	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}
