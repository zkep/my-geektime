package cli

import (
	"context"
	"os"
)

type Flags struct {
	Config string `name:"config" default:"config.yml" help:"Path to config file"`
}

type App struct {
	ctx  context.Context
	quit <-chan os.Signal
}

func NewApp(ctx context.Context, quit <-chan os.Signal) *App {
	return &App{ctx: ctx, quit: quit}
}

func (app *App) Run(_ *Flags) error {
	return nil
}
