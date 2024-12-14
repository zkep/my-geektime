package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/uaxe/cliz"
	"github.com/zkep/mygeektime/cmd/api"
	"github.com/zkep/mygeektime/cmd/cli"
	"github.com/zkep/mygeektime/lib/color"
)

type App struct {
	ctx    context.Context
	quit   <-chan os.Signal
	assets embed.FS
}

var (
	version = "0.0.1"
)

func banner(_ *cliz.Cli) string {
	return fmt.Sprintf("%s %s",
		color.Green("My GeekTime CLI"),
		color.Cyan(version),
	)
}

func Execute(assets embed.FS) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	NewApp(ctx, quit, assets).Run()
}

func NewApp(ctx context.Context, quit <-chan os.Signal, assets embed.FS) *App {
	return &App{ctx, quit, assets}
}

func (app *App) Context() context.Context {
	return app.ctx
}

func (app *App) Run() {
	c := cliz.NewCli("MyGeekTime", "The Go My GeekTime Server", version)

	c.SetBannerFunction(banner)

	c.PreRun(func(_ *cliz.Cli) error { return app.docter() })

	apiApp := api.NewApp(app.ctx, app.quit, app.assets)
	c.NewSubCommandFunction("server", "This is http server", apiApp.Run)

	cliApp := cli.NewApp(app.ctx, app.quit, app.assets)
	subCLI := c.NewSubCommand("cli", "This is command")
	subCLI.NewSubCommandFunction("browser", "install browser dependencies", cliApp.Browser)
	subCLI.NewSubCommandFunction("config", "generate config file templete", cliApp.Config)

	if err := c.Run(); err != nil {
		fmt.Println(color.Red(err.Error()))
	}
}

func (app *App) docter() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("Please install ffmpeg: ")
		fmt.Println("Ffmpeg will be used for video merging")
		fmt.Println()
		fmt.Println(color.Blue("https://ffmpeg.org/download.html"))
		fmt.Println()
		return err
	}
	if _, err := os.Stat("cookie.txt"); err != nil {
		if os.IsNotExist(err) {
			chromedriver := "./chromedriver"
			if runtime.GOOS == "windows" {
				chromedriver = "./chromedriver.exe"
			}
			if _, err1 := exec.LookPath(chromedriver); err1 != nil {
				fmt.Println("Please install chromedriver: ")
				fmt.Println("Chromedriver will be used by default to simulate login and obtain cookies")
				fmt.Println(color.Blue("https://ffmpeg.org/download.html"))
				fmt.Println()
				fmt.Println(color.Blue("Also you can save Geektime's cookie to 'cookie.txt' in current folder"))
				return fmt.Errorf("%w OR %w", err, err1)
			}
			return nil
		}
		return err
	}
	return nil
}
