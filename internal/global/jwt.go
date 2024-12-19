package global

import "github.com/zkep/mygeektime/lib/rest"

const (
	Identity    = "identity"
	Role        = "role"
	AccessToken = "access_token"
	Analogjwt   = "analogjwt"
)

var (
	JWT *rest.JWTConfig
)
