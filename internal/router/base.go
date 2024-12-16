package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func base(public, _ *gin.RouterGroup) {
	api := v2.NewBase()
	{
		public.POST("/base/login", api.Login)
		public.GET("/base/redirect", api.Redirect)
	}
}
