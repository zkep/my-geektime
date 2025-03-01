package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/my-geektime/internal/api/v2"
)

func file(public, _ *gin.RouterGroup) {
	api := v2.NewFile()
	{
		public.GET("/file/proxy", api.Proxy)
	}
}
