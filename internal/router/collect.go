package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/my-geektime/internal/api/v2"
)

func collect(_, private *gin.RouterGroup) {
	api := v2.NewCollect()
	{
		private.GET("/collect/list", api.List)
		private.POST("/collect/create", api.Create)
		private.DELETE("/collect/delete", api.Delete)
	}
}
