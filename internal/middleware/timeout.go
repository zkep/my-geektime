package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func TimeoutResponse(c *gin.Context) {
	c.JSON(
		http.StatusRequestTimeout,
		gin.H{"status": http.StatusRequestTimeout, "msg": "request timeout"},
	)
}

func Timeout() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(30*time.Second),
		timeout.WithResponse(TimeoutResponse),
	)
}
