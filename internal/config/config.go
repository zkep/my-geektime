package config

type Config struct {
	Server  Server   `json:"server" yaml:"server"`
	JWT     JWT      `json:"jwt"  yaml:"jwt"`
	I18N    I18N     `json:"i18n"  yaml:"i18n"`
	DB      Database `json:"database" yaml:"database"`
	Storage Storage  `json:"storage" yaml:"storage"`
	Browser Browser  `json:"browser" yaml:"browser"`
	Site    Site     `json:"site" yaml:"site"`
}
