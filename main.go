package main

import (
	"embed"

	"github.com/zkep/mygeektime/cmd"
)

//go:embed i18n/*
//go:embed web/index.html
//go:embed web/public/*
//go:embed web/pages/*
//go:embed config.yml
var Assets embed.FS

func main() { cmd.Execute(Assets) }
