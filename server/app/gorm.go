package app

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type db struct {
	cnn *gorm.DB
}

func (s *Server) newDb() *db {
	var err error

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", s.c.Db.User, s.c.Db.Pwd, s.c.Db.Host, s.c.Db.Port, s.c.Db.Name)

	conn, err := gorm.Open(mysql.Open(connection), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return &db{conn}
}
