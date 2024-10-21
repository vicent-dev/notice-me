package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	ID            uuid.UUID  `gorm:"type:uuid"`
	Body          string     `gorm:"body" json:"Body"`
	NotifiedAt    *time.Time `gorm:"notified_at" json:"NotifiedAt"`
	ClientId      string     `gorm:"client_id" json:"ClientId"`
	ClientGroupId string     `gorm:"client_group_id" json:"ClientGroupId"`
}

func NewNotification(body, clientId, clientGroupId string) *Notification {
	return &Notification{
		ID:            uuid.New(),
		Body:          body,
		ClientId:      clientId,
		ClientGroupId: clientGroupId,
	}
}

type NotificationPostDto struct {
	Body          string `json:"body"`
	ClientId      string `json:"clientId"`
	ClientGroupId string `json:"clientGroupId"`
}

func (n *Notification) FormatHTML() string {
	return n.Body + " <br/><br/> <small>" +
		"Created at: " + n.CreatedAt.Format(time.RFC3339) + " <br/> " +
		"Notified at: " + n.NotifiedAt.Format(time.RFC3339) + "</small>"
}
