package main

import (
	"database/sql"
	"fmt"
	"github.com/en-vee/alog"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func connectDb(c *config) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.db.user,
		c.db.pwd,
		c.db.host,
		c.db.port,
		c.db.name,
	))

	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getNotificationsPending(db *sql.DB) []*Notification {
	rows, err := db.Query("SELECT id FROM notifications WHERE notified_at = \"0000-00-00 00:00:00.000\"")
	var notifications []*Notification

	if err != nil {
		alog.Error(err.Error())
		return notifications
	}

	defer rows.Close()

	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID); err != nil {
			alog.Error(err.Error())
		}

		notifications = append(notifications, &n)
	}

	return notifications
}
