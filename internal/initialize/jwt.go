package initialize

import (
	"context"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/lib/rest"
)

func Jwt(_ context.Context) error {
	global.JWT = rest.JWT(global.CONF.JWT.Secret, global.CONF.JWT.Expires)
	return nil
}
