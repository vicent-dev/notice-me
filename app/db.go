package app

import (
	"fmt"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB connects to the database using the provided config, runs auto-migration,
// and returns the *gorm.DB handle. Callers must check the error.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Db.User, cfg.Db.Pwd, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)

	db, err := gorm.Open(mysql.Open(connection), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(&notification.Notification{}, &auth.ApiKey{}); err != nil {
		return nil, err
	}

	db.Exec("ALTER DATABASE " + cfg.Db.Name + " character set utf8mb4 collate utf8mb4_unicode_ci;")

	return db, nil
}

// InitRepositories creates the standard repository map (auth + notification)
// backed by the provided *gorm.DB. This is the lightweight initialization path
// for tools (like the CLI) that do not need the full HTTP/RabbitMQ stack.
func InitRepositories(db *gorm.DB) map[string]interface{} {
	repos := make(map[string]interface{})
	repos[notification.RepositoryKey] = repository.NewGorm[notification.Notification](db)
	repos[auth.RepositoryKey] = repository.NewGorm[auth.ApiKey](db)
	return repos
}

func (s *server) connectDb() {
	db, err := InitDB(s.c)
	if err != nil {
		panic(err)
	}
	s.db = db
}

func (s *server) initialiseRepositories() {
	if s.db == nil {
		s.connectDb()
	}
	s.repositories = InitRepositories(s.db)
}

func (s *server) getRepository(name string) interface{} {
	if r, ok := s.repositories[name]; ok {
		return r
	}

	return nil
}
