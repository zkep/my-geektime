package initialize

import (
	"context"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/storage"
)

func Storage(_ context.Context) error {
	s, err := storage.NewLocalStorage(global.CONF.Storage.Source)
	if err != nil {
		return err
	}
	global.Storage = s
	return nil
}
