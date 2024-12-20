package config

type Site struct {
	EmailRegisterSubject string `json:"email_register_subject" yaml:"email_register_subject"`
	EmailRegisterBody    string `json:"email_register_body" yaml:"email_register_body"`
	EmailRegisterAttach  string `json:"email_register_attach" yaml:"email_register_attach"`
}
