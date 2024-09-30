package notification

import (
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	gorm.Model
	Body          string    `gorm:"body" json:"Body"`
	NotifiedAt    time.Time `gorm:"notified_at" json:"NotifiedAt"`
	ClientId      string    `gorm:"client_id" json:"ClientId"`
	ClientGroupId string    `gorm:"client_group_id" json:"ClientGroupId"`
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
