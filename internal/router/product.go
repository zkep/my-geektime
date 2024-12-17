package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func product(_, private *gin.RouterGroup) {
	api := v2.NewProduct()
	{
		private.GET("/product/list", api.List)
		private.GET("/product/pvip/list", api.PvipProductList)
		private.GET("/product/articles", api.Articles)
		private.GET("/product/article/info", api.ArticleInfo)
		private.POST("/product/download", api.Download)
	}
}
