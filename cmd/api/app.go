package api

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/zkep/mygeektime/internal/config"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/initialize"
	"github.com/zkep/mygeektime/internal/router"
	"github.com/zkep/mygeektime/lib/browser"
	"github.com/zkep/mygeektime/lib/color"
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
	if err := initialize.InitRedis(app.ctx); err != nil {
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
	if err := app.docterFfmpeg(); err != nil {
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
	openURL := fmt.Sprintf("http://%s", addr)
	if f.Browser.OpenBrowser {
		_ = browser.Open(openURL)
	}
	fmt.Printf("browser open: %s\n", openURL)
	return nil
}

func (app *App) docterFfmpeg() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("Please install ffmpeg: ")
		fmt.Println("Ffmpeg will be used for video merging")
		fmt.Println()
		fmt.Println(color.Blue("https://ffmpeg.org/download.html"))
		fmt.Println()
		return err
	}
	return nil
}
