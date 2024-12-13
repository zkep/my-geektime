package config

type Storage struct {
	Driver    string `json:"driver" yaml:"driver"`
	Directory string `json:"directory" yaml:"directory"`
	Bucket    string `json:"bucket" yaml:"bucket"`
	Host      string `json:"host" yaml:"host"`
}
