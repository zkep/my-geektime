package initialize

import (
	"context"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/pool"
)

func GPool(ctx context.Context) error {
	global.GPool = pool.NewLimitPool(ctx, 10)
	return nil
}
