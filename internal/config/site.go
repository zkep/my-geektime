package config

type Site struct {
	Register Register `json:"register" yaml:"register"`
	Login    Login    `json:"login" yaml:"login"`
	Download bool     `json:"download" yaml:"download"`
}

type Register struct {
	Type  string        `json:"type" yaml:"type"`
	Email RegisterEmail `json:"email" yaml:"email"`
}

type Login struct {
	Type  string     `json:"type" yaml:"type"`
	Guest LoginGuest `json:"guest" yaml:"guest"`
}

type RegisterEmail struct {
	Subject string `json:"subject" yaml:"subject"`
	Body    string `json:"body" yaml:"body"`
	Attach  string `json:"attach" yaml:"attach"`
}

type LoginGuest struct {
	Name     string `json:"name" yaml:"name"`
	Password string `json:"password" yaml:"password"`
}
