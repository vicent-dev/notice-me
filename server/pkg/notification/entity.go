package notification

import "time"

type Notification struct {
	text       string
	notifiedAt time.Time
}
