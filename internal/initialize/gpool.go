package initialize

import (
	"context"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/pool"
)

func GPool(ctx context.Context) error {
	// Please keep one, otherwise there is danger in download
	global.GPool = pool.NewLimitPool(ctx, 1)
	return nil
}
