package db

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func MaxIdleConns(n int) func(opt *DbConfig) {
	return func(opt *DbConfig) {
		opt.MaxIdleConns = n
	}
}

func MaxOpenConns(n int) func(opt *DbConfig) {
	return func(opt *DbConfig) {
		opt.MaxOpenConns = n
	}
}

func ConnMaxLifetime(d time.Duration) func(opt *DbConfig) {
	return func(opt *DbConfig) {
		opt.ConnMaxLifetime = d
	}
}

func NewGORM(driver, source string, opts ...func(*DbConfig)) func() (*gorm.DB, error) {
	return func() (*gorm.DB, error) {
		var dialector gorm.Dialector
		switch driver {
		case "mysql":
			dialector = mysql.Open(source)
		case "postgres":
			dialector = postgres.Open(source)
		default:
			dialector = sqlite.Open(source)
		}

		db, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			return nil, err
		}
		cfg := &DbConfig{
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
		}
		for _, opt := range opts {
			opt(cfg)
		}
		tx, err := db.DB()
		if err != nil {
			return nil, err
		}
		tx.SetMaxIdleConns(cfg.MaxIdleConns)
		tx.SetMaxOpenConns(cfg.MaxOpenConns)
		tx.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		return db, nil
	}
}
