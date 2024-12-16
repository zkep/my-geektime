package config

type Browser struct {
	DriverPath  string `json:"driver_path" yaml:"driver_path"`
	OpenBrowser bool   `json:"open_browser" yaml:"open_browser"`
}
