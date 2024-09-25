package notification

import "time"

type Notification struct {
	Body       string    `json:"body"`
	NotifiedAt time.Time `json:"notified_at"`
}
