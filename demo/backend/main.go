package main

import (
	"fmt"
	"github.com/en-vee/alog"
	"github.com/joho/godotenv"
	"log"
)

// This demo service will fetch notifications and send them to the notify queue.
// In a real case scenario that would be handled by other backend service and publish the
// notification task after X business logic process finishes.

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := connectDb()
	rabbit := connectRabbit()

	defer db.Close()
	defer rabbit.Close()

	notifications := getNotificationsPending(db)

	for _, n := range notifications {
		publishNotificationNotify(rabbit, n)
	}

	alog.Info(fmt.Sprintf("%v notifications sent to notify to notice-me service", len(notifications)))
}
