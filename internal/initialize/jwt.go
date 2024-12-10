package initialize

import (
	"context"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/rest"
)

func Jwt(_ context.Context) error {
	global.JWT = rest.JWT(global.CONF.JWT.Secret, global.CONF.JWT.Expires)
	return nil
}
