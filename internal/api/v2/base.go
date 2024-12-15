package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/zkep/mygeektime/lib/browser"
	"github.com/zkep/mygeektime/lib/color"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
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
	cookiePath, err := filepath.Abs(global.CONF.Browser.CookiePath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusBadRequest,
			"msg":    err.Error(),
		})
	}
	global.CONF.Browser.CookiePath = cookiePath
	if stat, err := os.Stat(global.CONF.Browser.CookiePath); err == nil && stat.Size() > 0 {
		raw, err := os.ReadFile(global.CONF.Browser.CookiePath)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
		}
		if err = auth(string(raw), cookieSavePath(string(raw), global.CONF.Browser.CookiePath)); err == nil {
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
				"msg":    "OK",
				"token":  token,
				"user":   global.GeekUser,
				"expire": expire.Format(time.RFC3339),
			})
			return
		}
	}
	switch r.Type {
	case geek.LoginWithUser:
		if err = loginWithAccount(r); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
	case geek.LoginWithCookie:
		if err = auth(r.Account, cookieSavePath(r.Account, global.CONF.Browser.CookiePath)); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
	default:
		if err = docterChromedriver(""); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": http.StatusBadRequest,
				"msg":    err.Error(),
			})
			return
		}
		go func() {
			if err = loginWithSimulate(); err != nil {
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
		"msg":    "OK",
		"token":  token,
		"user":   global.GeekUser,
		"expire": expire.Format(time.RFC3339),
	})
}

const (
	ticketLoginURL = "https://account.geekbang.org/account/ticket/login"
	loginURL       = "https://account.geekbang.org/login"
	authURL        = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL     = "https://time.geekbang.org/dashboard/usercenter"
	geekTimeURL    = "https://time.geekbang.org"
)

func cookieSavePath(cookies, path string) func(r *http.Response) error {
	return func(r *http.Response) error {
		var auth geek.AuthResponse
		if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(cookies), os.ModePerm); err != nil {
			return err
		}
		global.GeekUser = auth.Data
		global.GeekCookies = cookies
		user := model.User{
			Uid:      fmt.Sprintf("%d", global.GeekUser.UID),
			NikeName: global.GeekUser.Nick,
			Avatar:   global.GeekUser.Avatar,
		}
		if err := global.DB.Where(
			model.User{
				Uid: fmt.Sprintf("%d", global.GeekUser.UID),
			}).
			Assign(model.User{
				UserName: global.GeekUser.Nick,
				Avatar:   global.GeekUser.Avatar,
			}).
			FirstOrCreate(&user).Error; err != nil {
			return err
		}
		return nil
	}
}

func auth(cookies string, after func(*http.Response) error) error {
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
		fmt.Println(color.Blue("Also you can save Geektime's cookie to 'cookie.txt' in current folder"))
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
		if err = auth(cookiesLine, cookieSavePath(cookiesLine, global.CONF.Browser.CookiePath)); err != nil {
			return false, err
		}

		addr := fmt.Sprintf("%s:%d", global.CONF.Server.HTTPAddr, global.CONF.Server.HTTPPort)
		openURL := fmt.Sprintf("http://%s", addr)
		_ = browser.Open(openURL)

		return true, nil
	}

	if err = wd.WaitWithTimeout(getCookiesCondition, time.Minute*5); nil != err {
		return err
	}
	return nil
}

func loginWithAccount(r base.LoginRequest) error {
	geek.DefaultLoginRequest.Cellphone = r.Account
	geek.DefaultLoginRequest.Password = r.Password
	loginData, _ := json.Marshal(geek.DefaultLoginRequest)
	err := zhttp.R.
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Referer", refererURL)
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Origin", "https://account.geekbang.org")
		}).
		After(func(r *http.Response) error {
			raw, _ := io.ReadAll(r.Body)
			fmt.Println(string(raw))
			r.Body = io.NopCloser(bytes.NewBuffer(raw))
			var l geek.LoginResponse
			if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
				return err
			}
			if l.Error.Msg != "" {
				return zhttp.BreakRetryError(errors.New(l.Error.Msg))
			}
			return nil
		}).
		DoWithRetry(context.Background(), http.MethodPost, ticketLoginURL, bytes.NewBuffer(loginData))
	if err != nil {
		return err
	}
	return nil
}
