package app

func (s *server) routes() {

	s.r.Use(loggingMiddleware)

	// websocket connection
	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	// api route group
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(jsonMiddleware)

	// notifications CRUD
	apiRouter.HandleFunc("/docs", s.docsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.createNotificationHandler()).Methods("POST")
	apiRouter.HandleFunc("/notifications", s.getNotificationsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.deleteNotificationHandler()).Methods("DELETE")
}
