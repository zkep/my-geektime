package db

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func NewRedis(host string, port int, password string) func() (*redis.Pool, error) {
	return func() (pool *redis.Pool, err error) {
		pool = &redis.Pool{
			MaxIdle:     16,
			MaxActive:   0,
			IdleTimeout: 300,
			Dial: func() (redis.Conn, error) {
				address := fmt.Sprintf(`%s:%d`, host, port)
				conn, er := redis.Dial("tcp", address)
				if er != nil {
					return nil, er
				}
				if password != "" {
					if _, er = conn.Do("AUTH", password); er != nil {
						return nil, conn.Close()
					}
				}
				return conn, nil
			},
		}
		// check conn
		c, err := pool.GetContext(context.Background())
		if err != nil {
			return
		}
		if err = c.Close(); err != nil {
			return
		}
		return
	}
}
