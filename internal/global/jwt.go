package global

import "github.com/zkep/mygeektime/lib/rest"

const (
	Identity    = "identity"
	AccessToken = "access_token"
	Analogjwt   = "analogjwt"
)

var (
	JWT *rest.JWTConfig
)
