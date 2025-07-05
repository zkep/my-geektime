package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/uaxe/cliz"
	"github.com/zkep/my-geektime/cmd/api"
	"github.com/zkep/my-geektime/cmd/cli"
	"github.com/zkep/my-geektime/libs/color"
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

	apiApp := api.NewApp(app.ctx, app.quit, app.assets)
	c.NewSubCommandFunction("server", "This is http server", apiApp.Run)

	cliApp := cli.NewApp(app.ctx, app.quit, app.assets)
	subCLI := c.NewSubCommand("cli", "This is command")
	subCLI.NewSubCommandFunction("config", "generate config file templete", cliApp.Config)
	subCLI.NewSubCommandFunction("data", "init geektime data", cliApp.Data)
	subCLI.NewSubCommandFunction("docs", "make geektime docs", cliApp.Docs)
	subCLI.NewSubCommandFunction("docs-local", "make geektime local docs", cliApp.LocalDoc)
	if err := c.Run(); err != nil {
		fmt.Println(color.Red(err.Error()))
	}
}
