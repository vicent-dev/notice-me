package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	"notice-me-server/static"
	"strconv"
	"time"

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

		notificationPostDto.ApiKeyValue = r.Header.Get(auth.API_KEY_HEADER)

		if notificationPostDto.Body == "" || notificationPostDto.ClientId == "" || notificationPostDto.ClientGroupId == "" {
			s.writeErrorResponse(w, errors.New("body, clientId and clientGroupId are required fields"), http.StatusBadRequest)
			return
		}

		n, err := notification.PublishCreateNotification(notificationPostDto, s.rabbit)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
		}

		s.writeResponse(w, n)
	}
}

func (s *server) getNotificationsHandler() func(w http.ResponseWriter, r *http.Request) {
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])

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
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		err := notification.DeleteNotification(id, repo)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		s.writeResponse(w, r)
	}
}

func (s *server) getNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		n, err := notification.GetNotification(id, repo)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		s.writeResponse(w, n)
	}
}

func (s *server) notifyNotificationHandler() func(w http.ResponseWriter, r *http.Request) {
	repo := s.getRepository(notification.RepositoryKey).(repository.Repository[notification.Notification])

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		n, err := notification.GetNotification(id, repo)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		if n.NotifiedAt != nil {
			err = errors.New(fmt.Sprintf("Notification already notified at: %s", n.NotifiedAt.Format(time.DateTime)))
			s.writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		err = notification.PublishNotifyNotification(n.ID.String(), s.rabbit)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		s.writeResponse(w, r)
	}
}

func (s *server) wsHandler() func(w http.ResponseWriter, r *http.Request) {
	ws := s.ws
	cors := s.c.Server.Cors

	return func(w http.ResponseWriter, r *http.Request) {
		hub.Upgrader.CheckOrigin = func(r *http.Request) bool {
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
			id = hub.AllClientId
		}

		if group == "" {
			group = hub.AllClientGroupId
		}

		conn, err := hub.Upgrader.Upgrade(w, r, nil)

		if err != nil {
			s.writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		client := hub.NewClient(
			id,
			group,
			ws,
			conn,
			make(chan []byte, 256),
		)

		client.WebsocketService.RegisterClient(client)

		go client.Write()
		go client.Read()
	}
}
