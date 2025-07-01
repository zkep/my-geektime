package global

import "github.com/zkep/my-geektime/internal/config"

var (
	CONF *config.Config

	CONFPath string
)

const (
	CustomConfigFile = "custom_config.yaml"
)
