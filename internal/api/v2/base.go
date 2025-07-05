package v2

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/base"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/user"
	"github.com/zkep/my-geektime/libs/utils"
	"gorm.io/gorm"
)

type Base struct{}

func NewBase() *Base {
	return &Base{}
}

func (b *Base) Login(c *gin.Context) {
	var r base.LoginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	var info model.User
	switch r.Type {
	case base.LoginWithName:
		var req base.LoginWithNameRequest
		if err := json.Unmarshal(r.Data, &req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if err := binding.Validator.ValidateStruct(req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if err := global.DB.
			Where(&model.User{
				UserName: req.Account,
				Status:   user.UserStatusActive,
			}).
			First(&info).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				global.FAIL(c, "base.login.error")
				return
			}
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if !utils.BcryptCheck(req.Password, info.Password) {
			global.FAIL(c, "base.login.error")
			return
		}
	default:
		global.FAIL(c, "base.login.type")
		return
	}
	token, expire, err := global.JWT.TokenGenerator(func(claims jwt.MapClaims) {
		claims[global.Identity] = info.Uid
		claims[global.Role] = info.RoleId
	})
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "OK",
		"token":  token,
		"user": base.User{
			Uid:      info.Uid,
			UserName: info.UserName,
			NickName: info.NickName,
			Avatar:   info.Avatar,
			GeekAuth: len(info.AccessToken) > 0,
			RoleId:   info.RoleId,
		},
		"expire": expire.Format(time.RFC3339),
	})
	c.SetCookie(global.Analogjwt, token, int(expire.Unix()), "/", "", false, false)
}

func (b *Base) Register(c *gin.Context) {
	var r base.RegisterRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	switch r.Type {
	case base.LoginWithName:
		var req base.RegisterWithNameRequest
		if err := json.Unmarshal(r.Data, &req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if err := binding.Validator.ValidateStruct(req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		var info model.User
		if err := global.DB.
			Where(&model.User{
				UserName: req.Account,
				Status:   user.UserStatusActive,
			}).
			First(&info).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if len(info.UserName) > 0 {
			global.FAIL(c, "base.register.exists")
			return
		}
		var count int64
		if err := global.DB.Model(&model.User{}).Count(&count).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		// first user is admin
		if count == 0 {
			info.RoleId = user.AdminRoleId
		}
		info.Uid = utils.HalfUUID()
		info.UserName = req.Account
		info.NickName = req.Account
		info.Password = utils.BcryptHash(req.Password)
		if err := global.DB.Create(&info).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
	default:
		global.FAIL(c, "base.register.type")
		return
	}
	global.OK(c, nil)
}

func (b *Base) RefreshCookie(c *gin.Context) {
	var r base.RefreshCookieRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	identity := c.GetString(global.Identity)
	var auth geek.AuthResponse
	if err := service.Authority(r.Cookie, service.SaveCookie(r.Cookie, identity, &auth)); err != nil {
		if errors.Is(err, service.ErrorGeekAccountNotLogin) {
			global.JSON(c, 10002, nil, "product.no_cookie", "")
		} else {
			global.FAIL(c, "fail.msg", err.Error())
		}
		return
	}
	global.OK(c, nil)
}

func (b *Base) Config(c *gin.Context) {
	ret := base.Config{
		RegisterType: global.CONF.Site.Register.Type,
		LoginType:    global.CONF.Site.Login.Type,
		LoginGuest: base.Guest{
			Name:     global.CONF.Site.Login.Guest.Name,
			Password: global.CONF.Site.Login.Guest.Password,
		},
	}
	global.OK(c, ret)
}
