package router

import (
	"github.com/gin-gonic/gin"
	v2 "github.com/zkep/my-geektime/internal/api/v2"
)

func task(public, private *gin.RouterGroup) {
	api := v2.NewTask()
	{
		private.GET("/task/list", api.List)
		private.GET("/task/info", api.Info)
		private.GET("/task/download", api.Download)
		private.DELETE("/task/delete", api.Delete)
		private.POST("/task/retry", api.Retry)
		private.GET("/task/export", api.Export)
		private.GET("/task/article/commonts", api.ArticleCommonts)
		private.GET("/task/article/discussions", api.ArticleDiscussion)
	}
	{
		public.GET("/task/kms", api.Kms)
		public.GET("/task/play.m3u8", api.Play)
		public.GET("/task/play/part", api.PlayPart)
	}
}
