package app

import (
	"fmt"
	"notice-me-server/pkg/notification"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (s *server) connectDb() {
	var err error

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", s.c.Db.User, s.c.Db.Pwd, s.c.Db.Host, s.c.Db.Port, s.c.Db.Name)

	conn, err := gorm.Open(mysql.Open(connection), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	s.db = conn

	s.db.AutoMigrate(&notification.Notification{})
}
