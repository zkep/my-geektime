package v2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/tebeka/selenium"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/base"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/user"
	"github.com/zkep/mygeektime/lib/browser"
	"github.com/zkep/mygeektime/lib/color"
	"github.com/zkep/mygeektime/lib/utils"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
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
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	var info model.User
	switch r.Type {
	case geek.AuthWithUser:
		if err := global.DB.
			Where(&model.User{
				UserName: r.Account,
				Status:   user.UserStatusActive,
			}).
			First(&info).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"status": http.StatusBadRequest,
					"msg":    "错误的用户名或密码",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": http.StatusBadRequest,
					"msg":    err.Error(),
				})
			}
			return
		}
		if !utils.BcryptCheck(r.Password, info.Password) {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    "错误的用户名或密码",
			})
			return
		}
	case geek.AuthWithCookie:
		var auth geek.AuthResponse
		if err := authority(r.Account, saveCookie(r.Account, &auth)); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
		info.Phone = auth.Data.Cellphone
		info.Avatar = auth.Data.Avatar
		info.NikeName = auth.Data.Nick
		info.Uid = fmt.Sprintf("%d", auth.Data.UID)
		info.AccessToken = r.Account
	default:
		if err := docterChromedriver(""); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
		go func() {
			if err := loginWithSimulate(); err != nil {
				global.LOG.Error("Login With Simulate", zap.Any("err", err))
			}
		}()
		c.JSON(http.StatusOK, gin.H{
			"status": 0,
			"msg":    "Browser Simulate Setup",
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
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
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
	var r base.RegisterRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Register Fail",
		})
		return
	}
	switch r.Type {
	case geek.AuthWithUser:
		var info model.User
		if err := global.DB.
			Where(&model.User{
				UserName: r.Account,
				Status:   user.UserStatusActive,
			}).
			First(&info).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
		if len(info.UserName) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    "当前账号已存在，请登录",
			})
			return
		}
		var count int64
		if err := global.DB.Model(&model.User{}).Count(&count).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
		// first user is admin
		if count == 0 {
			info.RoleId = 1
		}
		info.Uid = utils.HalfUUID()
		info.UserName = r.Account
		info.NikeName = r.Account
		info.Password = utils.BcryptHash(r.Password)
		if err := global.DB.Create(&info).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Unsupported registration type",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK"})
}

func (b *Base) Redirect(c *gin.Context) {
	var r base.RedirectRequest
	if err := c.ShouldBind(&r); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	tokenByte, err := base64.URLEncoding.DecodeString(r.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	cookieToken := string(tokenByte)
	var auth geek.AuthResponse
	if err = authority(cookieToken, saveCookie(cookieToken, &auth)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    err.Error(),
		})
		return
	}
	token, expire, err := global.JWT.DefaultTokenGenerator(
		func() (jwt.MapClaims, error) {
			claims := jwt.MapClaims{}
			claims[global.Identity] = auth.Data.UID
			claims[global.AccessToken] = cookieToken
			return claims, nil
		})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Login Fail",
		})
		return
	}
	c.SetCookie(global.Analogjwt, token, int(expire.Unix()), "/", "", false, false)
	c.Redirect(http.StatusFound, "/")
}

const (
	loginURL    = "https://account.geekbang.org/login"
	authURL     = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL  = "https://time.geekbang.org/dashboard/usercenter"
	geekTimeURL = "https://time.geekbang.org"
)

func saveCookie(cookies string, auth *geek.AuthResponse) func(r *http.Response) error {
	return func(r *http.Response) error {
		if err := json.NewDecoder(r.Body).Decode(auth); err != nil {
			return err
		}
		user := model.User{
			Uid:         fmt.Sprintf("%d", auth.Data.UID),
			NikeName:    auth.Data.Nick,
			Avatar:      auth.Data.Avatar,
			AccessToken: cookies,
			Phone:       auth.Data.Cellphone,
		}
		if err := global.DB.Where(
			model.User{
				Uid: fmt.Sprintf("%d", auth.Data.UID),
			}).
			Assign(model.User{
				UserName:    auth.Data.Nick,
				Avatar:      auth.Data.Avatar,
				AccessToken: cookies,
				Phone:       auth.Data.Cellphone,
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

func docterChromedriver(chromedriver string) error {
	if chromedriver == "" {
		driverPath, err := filepath.Abs(global.CONF.Browser.DriverPath)
		if err != nil {
			return err
		}
		chromedriver = driverPath
	}
	if runtime.GOOS == "windows" {
		chromedriver = fmt.Sprintf("%s.exe", strings.TrimSuffix(chromedriver, ".exe"))
	}
	if _, err := exec.LookPath(chromedriver); err != nil {
		fmt.Println("Please install chromedriver: ")
		fmt.Println("Chromedriver will be used by default to simulate login and obtain cookies")
		fmt.Println(color.Blue("https://googlechromelabs.github.io/chrome-for-testing/#stable"))
		fmt.Println()
		return err
	}
	return nil
}

func loginWithSimulate() error {
	driverPath, err := filepath.Abs(global.CONF.Browser.DriverPath)
	if err != nil {
		return err
	}
	if err = docterChromedriver(driverPath); err != nil {
		return err
	}
	global.CONF.Browser.DriverPath = driverPath
	port, err := zhttp.PickUnusedPort()
	if err != nil {
		return err
	}
	opts := []selenium.ServiceOption{
		selenium.Output(os.Stderr),
	}
	selenium.SetDebug(true)
	if runtime.GOOS == "windows" {
		global.CONF.Browser.DriverPath = strings.TrimSuffix(global.CONF.Browser.DriverPath, ".exe")
		global.CONF.Browser.DriverPath = fmt.Sprintf("%s.exe", global.CONF.Browser.DriverPath)
	}
	service, err := selenium.NewChromeDriverService(global.CONF.Browser.DriverPath, port, opts...)
	if err != nil {
		return err
	}
	defer func() { _ = service.Stop() }()
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return err
	}
	defer func() { _ = wd.Quit() }()

	if err = wd.Get(loginURL); nil != err {
		return err
	}

	getCookiesCondition := func(wd selenium.WebDriver) (bool, error) {
		currentURL, err := wd.CurrentURL()
		if err != nil {
			return false, err
		}
		if strings.Contains(currentURL, loginURL) {
			return false, nil
		}
		noLoop := strings.HasPrefix(currentURL, geekTimeURL)
		if !noLoop {
			return false, nil
		}
		cookies, err := wd.GetCookies()
		if err != nil {
			return false, err
		}
		cookiesLine := ""
		for k, v := range cookies {
			cookiesLine += fmt.Sprintf("%s=%s", v.Name, v.Value)
			if k < len(cookies)-1 {
				cookiesLine += ";"
			}
		}
		var auth geek.AuthResponse
		if err = authority(cookiesLine, saveCookie(cookiesLine, &auth)); err != nil {
			return false, err
		}
		addr := fmt.Sprintf("%s:%d", global.CONF.Server.HTTPAddr, global.CONF.Server.HTTPPort)
		token := base64.URLEncoding.EncodeToString([]byte(cookiesLine))
		openURL := fmt.Sprintf("http://%s/%s?token=%s", addr, "v2/base/redirect", token)
		_ = browser.Open(openURL)
		return true, nil
	}

	if err = wd.WaitWithTimeout(getCookiesCondition, time.Minute*5); nil != err {
		return err
	}
	return nil
}
