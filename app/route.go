package app

func (s *server) routes() {

	s.r.Use(s.loggingMiddleware)

	// websocket connection
	s.r.HandleFunc("/ws", s.wsHandler()).Methods("GET")

	// docs (no auth required)
	s.r.HandleFunc("/api/docs", s.docsHandler()).Methods("GET")

	// api route group
	apiRouter := s.r.PathPrefix("/api").Subrouter()
	apiRouter.Use(s.jsonMiddleware)
	apiRouter.Use(s.authMiddleware)

	// auth key management
	apiRouter.HandleFunc("/auth/keys", s.listKeysHandler()).Methods("GET")
	apiRouter.HandleFunc("/auth/keys/{id}", s.revokeKeyHandler()).Methods("DELETE")

	// notifications CRUD
	apiRouter.HandleFunc("/notifications", s.createNotificationHandler()).Methods("POST")
	apiRouter.HandleFunc("/notifications/notify/{id}", s.notifyNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications", s.getNotificationsHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.getNotificationHandler()).Methods("GET")
	apiRouter.HandleFunc("/notifications/{id}", s.deleteNotificationHandler()).Methods("DELETE")
}
