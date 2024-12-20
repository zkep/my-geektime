package v2

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/base"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/user"
	"github.com/zkep/mygeektime/lib/utils"
	"github.com/zkep/mygeektime/lib/zhttp"
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
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
		return
	}
	var info model.User
	switch r.Type {
	case geek.AuthWithUser:
		if err := global.DB.
			Where(&model.User{
				Email:  r.Email,
				Status: user.UserStatusActive,
			}).
			First(&info).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"status": http.StatusInternalServerError,
					"msg":    "错误的用户名或密码",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": http.StatusInternalServerError,
					"msg":    err.Error(),
				})
			}
			return
		}
		if !utils.BcryptCheck(r.Password, info.Password) {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "错误的用户名或密码",
			})
			return
		}
	default:
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "错误的登录方式",
		})
		return
	}
	token, expire, err := global.JWT.DefaultTokenGenerator(
		func() (jwt.MapClaims, error) {
			claims := jwt.MapClaims{}
			claims[global.Identity] = info.Uid
			claims[global.Role] = info.RoleId
			claims[global.AccessToken] = info.AccessToken
			return claims, nil
		})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
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
	var req base.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Register Fail",
		})
		return
	}
	switch req.Type {
	case geek.AuthWithUser:
		code, err := global.Redis.Get(c, req.Email).Result()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    err.Error(),
			})
			return
		}
		if !strings.EqualFold(code, req.Code) {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "验证码不正确",
			})
			return
		}
		var info model.User
		if err = global.DB.
			Where(&model.User{
				Email:  req.Email,
				Status: user.UserStatusActive,
			}).
			First(&info).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    err.Error(),
			})
			return
		}
		if len(info.UserName) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    "当前账号已存在，请登录",
			})
			return
		}
		var count int64
		if err := global.DB.Model(&model.User{}).Count(&count).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    err.Error(),
			})
			return
		}
		// first user is admin
		if count == 0 {
			info.RoleId = 1
		}
		info.Uid = utils.HalfUUID()
		info.Email = req.Email
		info.UserName = req.Email
		info.NikeName = req.Email
		info.Password = utils.BcryptHash(req.Password)
		if err := global.DB.Create(&info).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Unsupported registration type",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK"})
}

func (b *Base) SendEmail(c *gin.Context) {
	var req base.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
		return
	}
	gen := utils.NewStrGenerator(utils.StrGeneratorWithChars(utils.SimpleChars))
	code := gen.Random(6)
	d := gomail.NewDialer(global.CONF.Email.Host,
		global.CONF.Email.Port, global.CONF.Email.User, global.CONF.Email.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	m := gomail.NewMessage()
	m.SetHeader("From", global.CONF.Email.From)
	m.SetHeader("To", req.Email)
	m.SetHeader("Subject", "我的极客时间邮箱验证码")
	m.SetBody("text/html",
		fmt.Sprintf("验证码： <b>%s</b> <br/><br/> <b>👏 扫下方微信二维码，欢迎加入技术交流群</b>", code))
	m.Attach(".mygeektime/Wechat.jpg")
	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
		return
	}
	if err := global.Redis.SetEx(c, req.Email, code, time.Minute*2).Err(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK"})
}

func (b *Base) RefreshCookie(c *gin.Context) {
	var r base.RefreshCookieRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    err.Error(),
		})
		return
	}
	identity := c.GetString(global.Identity)
	var auth geek.AuthResponse
	if err := authority(r.Cookie, saveCookie(r.Cookie, identity, &auth)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK"})
}

const (
	authURL    = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL = "https://time.geekbang.org/dashboard/usercenter"
)

func saveCookie(cookies string, identity string, auth *geek.AuthResponse) func(r *http.Response) error {
	return func(r *http.Response) error {
		if err := json.NewDecoder(r.Body).Decode(auth); err != nil {
			return err
		}
		user := model.User{
			Uid:         identity,
			NikeName:    auth.Data.Nick,
			Avatar:      auth.Data.Avatar,
			AccessToken: cookies,
		}
		if err := global.DB.Where(model.User{Uid: identity}).
			Assign(model.User{
				UserName:    auth.Data.Nick,
				Avatar:      auth.Data.Avatar,
				AccessToken: cookies,
			}).
			FirstOrCreate(&user).Error; err != nil {
			return err
		}
		return nil
	}
}

func authority(cookies string, after func(*http.Response) error) error {
	jar, _ := cookiejar.New(nil)
	global.HttpClient = &http.Client{Jar: jar, Timeout: 5 * time.Minute}
	t := time.Now().UnixMilli()
	authUrl := fmt.Sprintf("%s?t=%d&v_t=%d", authURL, t, t)

	err := zhttp.R.Client(global.HttpClient).
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Referer", refererURL)
			r.Header.Set("Cookie", cookies)
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Origin", "https://time.geekbang.com")
		}).
		After(after).
		DoWithRetry(context.Background(), http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	return nil
}
