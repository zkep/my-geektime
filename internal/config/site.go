package config

type Site struct {
	Register Register `json:"register" yaml:"register"`
}

type Register struct {
	Type  string        `json:"type" yaml:"type"`
	Email RegisterEmail `json:"email" yaml:"email"`
}

type RegisterEmail struct {
	Subject string `json:"subject" yaml:"subject"`
	Body    string `json:"body" yaml:"body"`
	Attach  string `json:"attach" yaml:"attach"`
}
