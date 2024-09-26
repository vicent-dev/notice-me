package notification

import (
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	gorm.Model
	Body       string    `gorm:"body" json:"Body"`
	NotifiedAt time.Time `gorm:"notified_at" json:"NotifiedAt"`
}

type NotificationPostDto struct {
	Body string `json:"body"`
}

func (n *Notification) FormatHTML() string {
	return n.Body + " <br/><br/> <small>" +
		"Created at: " + n.CreatedAt.Format(time.RFC3339) + " <br/> " +
		"Notified at: " + n.NotifiedAt.Format(time.RFC3339) + "</small>"
}
