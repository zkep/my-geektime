package config

type Database struct {
	Driver       string `json:"driver" yaml:"driver"`
	Source       string `json:"source" yaml:"source"`
	MaxIdleConns int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"`
	Log          Zap    `json:"log" yaml:"log"`
}
