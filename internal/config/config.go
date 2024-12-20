package config

import "github.com/zkep/mygeektime/lib/zoauth"

type Config struct {
	Server  Server          `json:"server" yaml:"server"`
	JWT     JWT             `json:"jwt" yaml:"jwt"`
	DB      Database        `json:"database" yaml:"database"`
	Storage Storage         `json:"storage" yaml:"storage"`
	Browser Browser         `json:"browser" yaml:"browser"`
	Oauth2  []zoauth.Config `json:"oauth2" yaml:"oauth2"`
	Email   Email           `json:"email" yaml:"email"`
	Redis   Redis           `json:"redis" yaml:"redis"`
}
