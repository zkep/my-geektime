package global

import "github.com/zkep/my-geektime/libs/rest"

const (
	Identity    = "identity"
	Role        = "role"
	AccessToken = "access_token"
	Analogjwt   = "analogjwt"
)

var (
	JWT *rest.JWTConfig
)
