package initialize

import (
	"context"

	"github.com/zkep/my-geektime/internal/global"
	"go.uber.org/zap"
)

func Logger(_ context.Context) error {
	l := zap.NewExample()
	global.LOG = l
	zap.ReplaceGlobals(global.LOG)
	return nil
}
