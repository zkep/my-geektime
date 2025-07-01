package initialize

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"path/filepath"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/handler/resource"
	"github.com/zkep/my-geektime/internal/service"
)

func Resource(ctx context.Context) error {
	cacheKeyFn := func(uri string) string {
		hash := md5.New()
		hash.Reset()
		hash.Write([]byte(uri))
		hashStr := hex.EncodeToString(hash.Sum(nil))
		cacheKey := filepath.Join(global.CONF.Site.Proxy.CachePrefix, hashStr)
		return cacheKey
	}
	global.Resource = resource.NewResource(ctx, 10, cacheKeyFn, service.PorxyMatch, global.Storage)
	return nil
}
