package cli

import (
	"context"
	"os"
)

type Flags struct {
	Pid int64   `name:"pid"  description:"product id"`
	Id  []int64 `name:"id"  description:"article id"`
	Dir string  `name:"dir" description:"output directory" default:"./"`
}

type App struct {
	ctx  context.Context
	quit <-chan os.Signal
}

func NewApp(ctx context.Context, quit <-chan os.Signal) *App {
	return &App{ctx: ctx, quit: quit}
}

func (app *App) Browser(f *BrowserFlags) error { return browser(f) }
