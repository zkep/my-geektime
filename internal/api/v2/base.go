package v2

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/types/base"
)

type Base struct {
}

func NewBase() *Base {
	return &Base{}
}

func (b *Base) Login(c *gin.Context) {
	var r base.LoginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	token, expire, err := global.JWT.DefaultTokenGenerator(
		func() (jwt.MapClaims, error) {
			claims := jwt.MapClaims{}
			claims["identity"] = global.GeekUser.UID
			return claims, nil
		})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"token":  token,
		"user":   global.GeekUser,
		"expire": expire.Format(time.RFC3339),
	})
}
