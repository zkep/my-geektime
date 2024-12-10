package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func task(_, private *gin.RouterGroup) {
	api := v2.NewTask()
	{
		private.GET("/task/list", api.List)
	}
}
