package api

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/initialize"
	"github.com/zkep/my-geektime/internal/router"
	"github.com/zkep/my-geektime/libs/browser"
	"github.com/zkep/my-geektime/libs/color"
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
	var (
		cfg            config.Config
		customConfPath = global.CustomConfigFile
		configRaw      []byte
		err            error
	)
	if f.Config == "" {
		configRaw, err = app.assets.ReadFile("config.yml")
	} else {
		customConfPath = filepath.Join(filepath.Dir(f.Config), global.CustomConfigFile)
		configRaw, err = os.ReadFile(f.Config)
	}
	if stat, err1 := os.Stat(customConfPath); err1 == nil && stat.Size() > 0 {
		configRaw, err = os.ReadFile(customConfPath)
	}
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(configRaw, &cfg); err != nil {
		return err
	}
	global.CONF = &cfg
	global.CONFPath = f.Config
	if err = initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err = initialize.Jwt(app.ctx); err != nil {
		return err
	}
	if err = initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err = initialize.Storage(app.ctx); err != nil {
		return err
	}
	if err = initialize.GPool(app.ctx); err != nil {
		return err
	}
	if err = initialize.Tw(app.ctx); err != nil {
		return err
	}
	if err = initialize.I18N(app.ctx, app.assets); err != nil {
		return err
	}
	if err = initialize.Resource(app.ctx); err != nil {
		return err
	}
	if err = app.newHttpServer(global.CONF); err != nil {
		return err
	}

	<-app.quit

	global.Resource.Stop()
	global.GPool.Close()

	return nil
}

func (app *App) newHttpServer(f *config.Config) error {
	if err := app.doctorFfmpeg(); err != nil {
		return err
	}
	if err := app.doctorMkdocs(); err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", f.Server.HTTPAddr, f.Server.HTTPPort)

	r, err := router.NewRouter(app.assets)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func (app *App) doctorFfmpeg() error {
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

func (app *App) doctorMkdocs() error {
	if _, err := exec.LookPath("mkdocs"); err != nil {
		fmt.Println("Please install mkdocs: ")
		fmt.Println("pip install mkdocs-material")
		fmt.Println()
		fmt.Println(color.Blue("https://github.com/mkdocs/mkdocs"))
		fmt.Println("install mkdocs-material, Please wait .....")
		name := "pip"
		if _, pipxErr := exec.LookPath("pipx"); pipxErr == nil {
			name = "pipx"
		}
		err = exec.CommandContext(app.ctx, name, "install", "mkdocs-material").Run()
		if err != nil {
			return err
		}
		fmt.Println("pip install mkdocs-material succeed")
	}
	return nil
}
