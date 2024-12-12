package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
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

	c.NewSubCommandFunction("server", "This is http server",
		api.NewApp(app.ctx, app.quit, app.assets).Run)

	cliApp := cli.NewApp(app.ctx, app.quit)
	c.NewSubCommand("cli", "This is Command").
		NewSubCommandFunction("browser", "install browser dependencies", cliApp.Browser)

	if err := c.Run(); err != nil {
		fmt.Println(color.Red(err.Error()))
	}
}
