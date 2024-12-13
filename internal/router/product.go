package router

import (
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func product(_, private *gin.RouterGroup) {
	store := persistence.NewInMemoryStore(time.Second)
	api := v2.NewProduct()
	{
		private.GET("/product/list", cache.CachePage(store, time.Second*5, api.List))
		private.GET("/product/articles", cache.CachePage(store, time.Minute, api.Articles))
		private.GET("/product/article/info", cache.CachePage(store, time.Minute, api.ArticleInfo))
		private.POST("/product/download", api.Download)
	}
}
