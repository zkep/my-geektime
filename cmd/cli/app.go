package cli

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
)

type Flags struct {
	Pid int64   `name:"pid"  description:"product id"`
	Id  []int64 `name:"id"  description:"article id"`
	Dir string  `name:"dir" description:"output directory" default:"./"`
}

type ConfigFlags struct {
	Config string `name:"config" default:"config_templete.yml" description:"generate config file"`
}

type App struct {
	ctx    context.Context
	quit   <-chan os.Signal
	assets embed.FS
}

func NewApp(ctx context.Context, quit <-chan os.Signal, assets embed.FS) *App {
	return &App{ctx, quit, assets}
}

func (app *App) Config(f *ConfigFlags) error {
	fi, err := app.assets.Open("config.yml")
	if err != nil {
		return err
	}
	defer func() { _ = fi.Close() }()
	fs, err := os.Create(f.Config)
	if err != nil {
		return err
	}
	defer func() { _ = fs.Close() }()
	_, err = io.Copy(fs, fi)
	if err != nil {
		return err
	}
	fmt.Printf("successfully created %s\n", f.Config)
	return nil
}
