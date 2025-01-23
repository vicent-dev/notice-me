package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const RepositoryKey = "notification"

type Notification struct {
	gorm.Model
	ID             uuid.UUID  `gorm:"type:uuid"`
	Body           string     `gorm:"body" json:"Body"`
	NotifiedAt     *time.Time `gorm:"notified_at" json:"NotifiedAt"`
	ClientId       string     `gorm:"client_id" json:"ClientId"`
	ClientGroupId  string     `gorm:"client_group_id" json:"ClientGroupId"`
	Instant        bool       `gorm:"instant;default:0" json:"Instant"`
	OriginClientId string     `json:"OriginClientId"`
}

func NewNotification(body, clientId, clientGroupId string, instant bool, originClientId string) *Notification {
	return &Notification{
		ID:             uuid.New(),
		Body:           body,
		ClientId:       clientId,
		ClientGroupId:  clientGroupId,
		Instant:        instant,
		OriginClientId: originClientId,
	}
}

type NotificationPostDto struct {
	Body           string `json:"body"`
	ClientId       string `json:"clientId"`
	ClientGroupId  string `json:"clientGroupId"`
	Instant        bool   `json:"instant"`
	OriginClientId string `json:"originClientId"`
}

type NotificationNotifyDto struct {
	ID string `json:"id"`
}

func (n *Notification) FormatHTML() string {
	return n.Body
}
