package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
	"github.com/zkep/mygeektime/internal/middleware"
)

func product(_, private *gin.RouterGroup) {
	api := v2.NewProduct()
	p := private.Group("/product", middleware.AccessToken())
	{
		p.GET("/pvip/list", api.PvipProductList)
		p.GET("/list", api.ProductList)
		p.GET("/articles", api.Articles)
		p.GET("/article/info", api.ArticleInfo)
		p.GET("/article/commonts", api.ArticleCommonts)
		p.GET("/article/discussions", api.ArticleDiscussion)
		p.POST("/download", api.Download)
	}
}
