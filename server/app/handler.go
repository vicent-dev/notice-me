package app

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/rabbit"
	"notice-me-server/pkg/repository"
	"notice-me-server/pkg/websocket"
	"notice-me-server/static"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *server) docsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		docsFile := static.GetDocsFile()

		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(docsFile)
	}
}

func (s *server) createNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	rbbt := rabbit.NewRabbit(s.amqp, s.c.Rabbit.ConsumersCount, s.c.Rabbit.Queues)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		notificationPostDto := &notification.NotificationPostDto{}
		if err = json.Unmarshal(body, notificationPostDto); err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		if notificationPostDto.Body == "" || notificationPostDto.ClientId == "" || notificationPostDto.ClientGroupId == "" {
			s.writeErrorResponse(w, errors.New("body, clientId and clientGroupId are required fields"), http.StatusBadRequest)
			return
		}

		n, err := notification.PublishCreateNotification(notificationPostDto, rbbt)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
		}

		s.writeResponse(w, n)
	}
}

func (s *server) getNotificationsHandler() func(w http.ResponseWriter, r *http.Request) {
	repo := repository.GetRepository[notification.Notification](s.db)

	return func(w http.ResponseWriter, r *http.Request) {
		pageSize := r.URL.Query().Get("pageSize")

		if pageSize == "" {
			pageSize = repository.DefaultPageSize
		}

		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		page := r.URL.Query().Get("page")
		if page == "" {
			page = repository.DefaultPage
		}

		pageInt, err := strconv.Atoi(page)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		ns, err := notification.GetNotifications(repo, pageSizeInt, pageInt)
		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		s.writeResponse(w, ns)
	}
}

func (s *server) deleteNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	repo := repository.GetRepository[notification.Notification](s.db)

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		err := notification.DeleteNotification(id, repo)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
		}

		s.writeResponse(w, r)
	}
}

func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws
	cors := s.c.Server.Cors

	return func(w http.ResponseWriter, r *http.Request) {
		websocket.Upgrader.CheckOrigin = func(r *http.Request) bool {
			for _, host := range cors {
				if host == r.Host {
					return true
				}

				if host == "*" {
					return true
				}
			}

			return false
		}

		id := r.URL.Query().Get("id")
		group := r.URL.Query().Get("groupId")

		if id == "" {
			id = websocket.AllClientId
		}

		if group == "" {
			group = websocket.AllClientGroupId
		}

		conn, err := websocket.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err)
			return
		}

		client := &websocket.Client{
			ID:               id,
			GroupId:          group,
			WebsocketService: ws,
			Conn:             conn,
			Send:             make(chan []byte, 256),
		}

		client.WebsocketService.Register <- client

		go client.Write()
		go client.Read()
	}
}
