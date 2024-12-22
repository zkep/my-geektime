package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/user"
)

func AccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var u model.User
		if err := global.DB.
			Where(&model.User{RoleId: user.AdminRoleId}).
			First(&u).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "no auth ! please refesh geektime cookie. ",
			})
			return
		}
		c.Set(global.AccessToken, u.AccessToken)
		c.Next()
	}
}
