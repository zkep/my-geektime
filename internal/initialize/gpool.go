package initialize

import (
	"context"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/lib/pool"
)

func GPool(ctx context.Context) error {
	global.GPool = pool.NewLimitPool(ctx, 10)
	return nil
}
