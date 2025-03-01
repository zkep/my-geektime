package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/my-geektime/internal/api/v2"
)

func user(_, private *gin.RouterGroup) {
	api := v2.NewUser()
	{
		private.GET("/user/list", api.List)
		private.POST("/user/status", api.Status)
	}
}
