package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/zkep/mygeektime/internal/config"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/initialize"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/router"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/lib/browser"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Flags struct {
	Config string `name:"config" description:"Path to config file"`
}

type App struct {
	ctx    context.Context
	quit   <-chan os.Signal
	assets embed.FS
}

func NewApp(ctx context.Context, quit <-chan os.Signal, assets embed.FS) *App {
	return &App{ctx, quit, assets}
}

func (app *App) Run(f *Flags) error {
	var cfg config.Config
	if f.Config == "" {
		fi, err := app.assets.Open("config.yml")
		if err != nil {
			return err
		}
		defer func() { _ = fi.Close() }()
		if err = yaml.NewDecoder(fi).Decode(&cfg); err != nil {
			return err
		}
	} else {
		fi, err := os.Open(f.Config)
		if err != nil {
			return err
		}
		defer func() { _ = fi.Close() }()
		if err = yaml.NewDecoder(fi).Decode(&cfg); err != nil {
			return err
		}
	}
	global.CONF = &cfg
	if err := initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err := initialize.Jwt(app.ctx); err != nil {
		return err
	}
	if err := initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err := initialize.Storage(app.ctx); err != nil {
		return err
	}
	if err := initialize.GPool(app.ctx); err != nil {
		return err
	}
	if err := initialize.Tw(app.ctx); err != nil {
		return err
	}

	options := []fx.Option{
		fx.Provide(func() context.Context { return app.ctx }),
		fx.Provide(func() *config.Config { return global.CONF }),
	}

	options = append(options,
		// register http handler
		fx.Invoke(app.newHttpServer),
	)

	depInj := fx.New(options...)
	if err := depInj.Start(app.ctx); err != nil {
		return err
	}

	<-app.quit

	return depInj.Stop(app.ctx)
}

func (app *App) newHttpServer(f *config.Config) error {
	if err := app.Login(); err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", f.Server.HTTPAddr, f.Server.HTTPPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router.NewRouter(app.assets),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.LOG.Error("listen: ", zap.Error(err))
		}
	}()
	if f.Browser.OpenBrowser {
		openURL := fmt.Sprintf("http://%s", addr)
		if err := browser.Open(openURL); err != nil {
			return err
		}
	}
	return nil
}

const (
	loginURL    = "https://account.geekbang.org/login"
	authURL     = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL  = "https://time.geekbang.org/dashboard/usercenter"
	geekTimeURL = "https://time.geekbang.org"
)

func (app *App) auth(cookies, path string) error {
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
		After(func(r *http.Response) error {
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
		}).
		DoWithRetry(context.Background(), http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) Login() error {
	cookiePath, err := filepath.Abs(global.CONF.Browser.CookiePath)
	if err != nil {
		return err
	}
	global.CONF.Browser.CookiePath = cookiePath
	driverPath, err := filepath.Abs(global.CONF.Browser.DriverPath)
	if err != nil {
		return err
	}
	global.CONF.Browser.DriverPath = driverPath
	if stat, err := os.Stat(global.CONF.Browser.CookiePath); err == nil && stat.Size() > 0 {
		if raw, err := os.ReadFile(global.CONF.Browser.CookiePath); err != nil {
			return err
		} else if err = app.auth(string(raw), global.CONF.Browser.CookiePath); err == nil {
			return nil
		}
	}
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
		if err = app.auth(cookiesLine, global.CONF.Browser.CookiePath); err != nil {
			return false, err
		}
		return true, nil
	}
	if err = wd.WaitWithTimeout(getCookiesCondition, time.Minute*5); nil != err {
		return err
	}
	return nil
}
