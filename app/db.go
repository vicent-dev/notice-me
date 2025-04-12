package app

import (
	"fmt"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (s *server) connectDb() {
	var err error

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", s.c.Db.User, s.c.Db.Pwd, s.c.Db.Host, s.c.Db.Port, s.c.Db.Name)

	conn, err := gorm.Open(mysql.Open(connection), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	sqlDB, _ := conn.DB()

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	s.db = conn

	err = s.db.AutoMigrate(&notification.Notification{})
	if err != nil {
		panic(err)
	}

	s.db.Exec("ALTER DATABASE " + s.c.Db.Name + " character set utf8mb4 collate utf8mb4_unicode_ci;")
}

func (s *server) initialiseRepositories() {
	if s.db == nil {
		s.connectDb()
	}

	s.repositories = make(map[string]interface{})

	s.repositories[notification.RepositoryKey] = repository.NewGorm[notification.Notification](s.db)
}

func (s *server) getRepository(name string) interface{} {
	if r, ok := s.repositories[name]; ok {
		return r
	}

	return nil
}
