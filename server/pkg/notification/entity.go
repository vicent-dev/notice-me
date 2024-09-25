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

func (n *Notification) Format() string {
	return n.Body + " - " +
		"Created at: " + n.CreatedAt.Format(time.RFC3339) + " - " +
		"Notified at: " + n.NotifiedAt.Format(time.RFC3339)
}
