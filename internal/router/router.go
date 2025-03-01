package router

import (
	"embed"
	"net/http"
	"path"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	mw "github.com/zkep/my-geektime/internal/middleware"
)

func NewRouter(assets embed.FS) *gin.Engine {
	e := gin.Default()

	e.Use(mw.Cors())

	e.Use(static.Serve("/", static.EmbedFolder(assets, "web")))

	e.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	if global.CONF.Storage.Driver == config.StorageLocal {
		e.StaticFS(path.Join("/", global.CONF.Storage.Bucket),
			gin.Dir(global.CONF.Storage.Directory, true))
	}

	public := e.Group("v2", mw.Timeout())
	private := e.Group("v2", mw.JWTMiddleware(), mw.Timeout())

	base(public, private)

	product(public, private)

	task(public, private)

	user(public, private)

	file(public, private)

	return e
}
