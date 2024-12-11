package cli

import (
	"context"
	"fmt"
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

func (app *App) Run(f *Flags) error {
	if len(f.Id) == 0 && f.Pid == 0 {
		return fmt.Errorf("no ids or no pid")
	}
	return nil
}
