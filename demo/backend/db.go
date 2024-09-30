package main

import (
	"database/sql"
	"fmt"
	"github.com/en-vee/alog"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

func connectDb() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_SCHEMA"),
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
