package config

type Redis struct {
	Addr         string `json:"addr" yaml:"addr"`
	Username     string `json:"username" yaml:"username"`
	Password     string `json:"password" yaml:"password"`
	PoolSize     int    `json:"pool_size" yaml:"pool_size"`
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"`
}
