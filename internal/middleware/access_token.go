package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/types/user"
)

func AccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var u model.User
		if err := global.DB.
			Where(&model.User{RoleId: user.AdminRoleId}).
			First(&u).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "no auth ! please refresh geektime cookie. ",
			})
			return
		}
		c.Set(global.AccessToken, u.AccessToken)
		c.Next()
	}
}
