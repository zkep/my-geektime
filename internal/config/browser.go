package config

type Browser struct {
	DriverPath  string `json:"driver_path" yaml:"driver_path"`
	CookiePath  string `json:"cookie_path" yaml:"cookie_path"`
	OpenBrowser bool   `json:"open_browser" yaml:"open_browser"`
}
