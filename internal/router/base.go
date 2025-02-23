package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func base(public, private *gin.RouterGroup) {
	api := v2.NewBase()
	{
		public.GET("/base/config", api.Config)
		public.POST("/base/login", api.Login)
		public.POST("/base/register", api.Register)
	}
	{
		private.POST("/base/refresh/cookie", api.RefreshCookie)
	}
}
