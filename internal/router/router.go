package router

import (
	"embed"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	mw "github.com/zkep/mygeektime/internal/middleware"
)

func NewRouter(assets embed.FS) *gin.Engine {
	e := gin.Default()

	e.Use(mw.Cors())

	e.Use(static.Serve("/", static.EmbedFolder(assets, "web")))

	e.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	public := e.Group("v2", mw.Timeout())
	private := e.Group("v2", mw.JWTMiddleware(), mw.Timeout())

	base(public, private)

	product(public, private)

	task(public, private)

	return e
}
