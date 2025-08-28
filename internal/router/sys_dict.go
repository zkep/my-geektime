package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/my-geektime/internal/api/v2"
)

func dict(_, private *gin.RouterGroup) {
	api := v2.NewDict()
	{
		private.GET("/sys/dict/tree", api.Tree)
		private.GET("/sys/dict/list", api.List)
		private.POST("/sys/dict/create", api.Create)
		private.PUT("/sys/dict/update", api.Update)
		private.DELETE("/sys/dict/delete", api.Delete)
	}
}
