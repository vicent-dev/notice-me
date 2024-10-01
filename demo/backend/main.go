package main

import (
	"fmt"
	"github.com/en-vee/alog"
)

// This demo service will fetch notifications and send them to the notify queue.
// In a real case scenario that would be handled by other backend service and publish the
// notification task after X business logic process finishes.

func main() {
	c := loadConfig()

	db := connectDb(c)
	rabbit := connectRabbit(c)

	defer db.Close()
	defer rabbit.Close()

	notifications := getNotificationsPending(db)

	for _, n := range notifications {
		publishNotificationNotify(rabbit, n)
	}

	alog.Info(fmt.Sprintf("%v notifications sent to notify to notice-me service", len(notifications)))
	deleteOldNotifications(db)
}
