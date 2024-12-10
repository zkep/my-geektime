package config

type JWT struct {
	Secret  string `json:"secret" yaml:"secret"`
	Expires int64  `json:"expires" yaml:"expires"`
}
