package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/base"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/user"
	"github.com/zkep/mygeektime/lib/utils"
	"gopkg.in/gomail.v2"
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
	case base.LoginWithEmail:
		var req base.LoginWithEmailRequest
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
				Email:  req.Email,
				Status: user.UserStatusActive,
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
		"user":   info,
		"role":   info.RoleId,
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
	case base.LoginWithEmail:
		var req base.RegisterWithEmailRequest
		if err := json.Unmarshal(r.Data, &req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if err := binding.Validator.ValidateStruct(req); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		code, err := global.Redis.Get(c, req.Email).Result()
		if err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		if !strings.EqualFold(code, req.Code) {
			global.FAIL(c, "base.register.err_code")
			return
		}
		var info model.User
		if err = global.DB.
			Where(&model.User{
				Email:  req.Email,
				Status: user.UserStatusActive,
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
		if err = global.DB.Model(&model.User{}).Count(&count).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		// first user is admin
		if count == 0 {
			info.RoleId = user.AdminRoleId
		}
		info.Uid = utils.HalfUUID()
		info.Email = req.Email
		info.UserName = req.Email
		info.NikeName = req.Email
		info.Password = utils.BcryptHash(req.Password)
		if err = global.DB.Create(&info).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
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
		info.NikeName = req.Account
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

func (b *Base) SendEmail(c *gin.Context) {
	var req base.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	gen := utils.NewStrGenerator(utils.StrGeneratorWithChars(utils.SimpleChars))
	code := gen.Random(6)
	d := gomail.NewDialer(global.CONF.Email.Host,
		global.CONF.Email.Port, global.CONF.Email.User, global.CONF.Email.Password)
	m := gomail.NewMessage()
	m.SetHeader("From", global.CONF.Email.From)
	m.SetHeader("To", req.Email)
	m.SetHeader("Subject", global.CONF.Site.Register.Email.Subject)
	m.SetBody("text/html", fmt.Sprintf(global.CONF.Site.Register.Email.Body, code))
	if global.CONF.Site.Register.Email.Attach != "" {
		m.Attach(global.CONF.Site.Register.Email.Attach)
	}
	if err := d.DialAndSend(m); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if err := global.Redis.SetEx(c, req.Email, code, time.Minute*2).Err(); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
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
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	global.OK(c, nil)
}

func (b *Base) Config(c *gin.Context) {
	ret := base.Config{
		RegisterType: global.CONF.Site.Register.Type,
	}
	global.OK(c, ret)
}
