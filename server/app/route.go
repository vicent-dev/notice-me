package app

func (s *server) routes() {

	s.r.Use(loggingMiddleware)

	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(jsonMiddleware)

	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	//create notification
	apiRouter.HandleFunc("/notification", s.createNotificationHandler()).Methods("POST")
}
