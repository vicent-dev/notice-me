package main

import (
	"database/sql"
	"fmt"
	"github.com/en-vee/alog"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

func connectDb(c *config) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.Db.User,
		c.Db.Pwd,
		c.Db.Host,
		c.Db.Port,
		c.Db.Name,
	))

	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getNotificationsPending(db *sql.DB) []*Notification {
	rows, err := db.Query("SELECT id FROM notifications WHERE notified_at IS NULL AND deleted_at IS NULL")
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

func deleteOldNotifications(db *sql.DB) {
	deleteCreatedAtTime := time.Now().Local().Add(time.Duration(-1) * time.Hour)

	rows, err := db.Exec("DELETE FROM notifications WHERE notified_at IS NOT NULL AND created_at < \"" + deleteCreatedAtTime.Format(time.DateTime) + "\"")

	if err != nil {
		alog.Error("Error trying to delete old notifications " + err.Error())
		return
	}

	rowsAffected, _ := rows.RowsAffected()

	alog.Info("Number of old (" + deleteCreatedAtTime.Format(time.DateTime) + ") notifications deleted: " + strconv.Itoa(int(rowsAffected)))
}
