package initialize

import (
	"context"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/libs/db"
)

func Gorm(_ context.Context) error {
	g, err := db.NewGORM(
		global.CONF.DB.Driver,
		global.CONF.DB.Source,
		db.MaxIdleConns(global.CONF.DB.MaxIdleConns),
		db.MaxOpenConns(global.CONF.DB.MaxOpenConns),
	)()
	if err != nil {
		return err
	}
	global.DB = g
	if err = g.AutoMigrate(
		&model.User{},
		&model.Task{},
		&model.Article{},
		&model.ArticleSimple{},
		&model.Product{},
		&model.ArticleComment{},
		&model.ArticleCommentDiscussion{},
	); err != nil {
		return err
	}
	return nil
}
