package config

type Server struct {
	AppName  string `yaml:"app_name"`
	RunMode  string `yaml:"run_mode"`
	HTTPAddr string `yaml:"http_addr"`
	HTTPPort int    `yaml:"http_port"`
}
