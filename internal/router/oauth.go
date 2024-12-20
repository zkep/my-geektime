package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/mygeektime/internal/api/v2"
)

func oauth(public, _ *gin.RouterGroup) {
	api := v2.NewOauth()
	{
		public.GET("/oauth2/authorize", api.Authorize)
	}
}
