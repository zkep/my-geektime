package config

type Site struct {
	Cache    bool     `json:"cache" yaml:"cache"`
	Download bool     `json:"download" yaml:"download"`
	Register Register `json:"register" yaml:"register"`
	Login    Login    `json:"login" yaml:"login"`
	Play     Play     `json:"play" yaml:"play"`
	Proxy    Proxy    `json:"proxy" yaml:"proxy"`
}

type (
	Register struct {
		Type  string        `json:"type" yaml:"type"`
		Email RegisterEmail `json:"email" yaml:"email"`
	}

	Login struct {
		Type  string     `json:"type" yaml:"type"`
		Guest LoginGuest `json:"guest" yaml:"guest"`
	}

	RegisterEmail struct {
		Subject string `json:"subject" yaml:"subject"`
		Body    string `json:"body" yaml:"body"`
		Attach  string `json:"attach" yaml:"attach"`
	}

	LoginGuest struct {
		Name     string `json:"name" yaml:"name"`
		Password string `json:"password" yaml:"password"`
	}

	Play struct {
		Type     string   `json:"type" yaml:"type"`
		ProxyUrl []string `json:"proxy_url" yaml:"proxy_url"`
	}

	Proxy struct {
		Urls        []string `json:"urls" yaml:"urls"`
		ProxyUrl    string   `json:"proxy_url" yaml:"proxy_url"`
		Cache       bool     `json:"cache" yaml:"cache"`
		CachePrefix string   `json:"cache_prefix" yaml:"cache_prefix"`
	}
)
