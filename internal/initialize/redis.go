package initialize

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zkep/mygeektime/internal/global"
)

func InitRedis(_ context.Context) error {
	opt := &redis.Options{
		Addr:         global.CONF.Redis.Addr,
		Username:     global.CONF.Redis.Username,
		Password:     global.CONF.Redis.Password,
		DB:           0,
		MaxRetries:   3,
		DialTimeout:  20 * time.Second,
		PoolSize:     global.CONF.Redis.PoolSize,
		MinIdleConns: global.CONF.Redis.MaxOpenConns,
	}
	global.Redis = redis.NewClient(opt)
	return global.Redis.Ping(context.Background()).Err()
}
