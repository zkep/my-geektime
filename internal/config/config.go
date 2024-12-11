package config

type Config struct {
	Server   Server   `json:"server" yaml:"server"`
	JWT      JWT      `json:"jwt" yaml:"jwt"`
	DB       Database `json:"database" yaml:"database"`
	Storage  Storage  `json:"storage" yaml:"storage"`
	Browser  Browser  `json:"browser" yaml:"browser"`
	Geektime Geektime `json:"geektime" yaml:"geektime"`
}
