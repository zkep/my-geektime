package geek

type AuthResponse struct {
	Error []any    `json:"error,omitempty"`
	Extra []any    `json:"extra,omitempty"`
	Data  GeekUser `json:"data,omitempty"`
	Code  int      `json:"code,omitempty"`
}

type GeekUser struct {
	Usersubtype   int    `json:"usersubtype,omitempty"`
	Usertype      int    `json:"usertype,omitempty"`
	Haspass       int    `json:"haspass,omitempty"`
	Pvip          int    `json:"pvip,omitempty"`
	Cestype       int    `json:"cestype,omitempty"`
	Fillnum       int    `json:"fillnum,omitempty"`
	Euid          string `json:"euid,omitempty"`
	Avatar        string `json:"avatar,omitempty"`
	Domain        string `json:"domain,omitempty"`
	SystemUpgrade int    `json:"system_upgrade,omitempty"`
	Medalid       int    `json:"medalid,omitempty"`
	Student       int    `json:"student,omitempty"`
	Cellphone     string `json:"cellphone,omitempty"`
	UID           int    `json:"uid,omitempty"`
	Cert          int    `json:"cert,omitempty"`
	Ctime         string `json:"ctime,omitempty"`
	Country       string `json:"country,omitempty"`
	Appid         int    `json:"appid,omitempty"`
	Cesid         string `json:"cesid,omitempty"`
	Nick          string `json:"nick,omitempty"`
}
