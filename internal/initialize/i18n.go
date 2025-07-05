package initialize

import (
	"context"
	"embed"
	"io/fs"
	"path"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/libs/rest"
	"go.uber.org/zap"
)

func I18N(_ context.Context, assets embed.FS) error {
	dirs, err := assets.ReadDir(global.CONF.I18N.Directory)
	if err != nil {
		return err
	}
	files := make([]fs.File, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			fpath := path.Join(global.CONF.I18N.Directory, dir.Name())
			file, er := assets.Open(fpath)
			if er != nil {
				return er
			}
			files = append(files, file)
		}
	}
	i18n, err := rest.InitI18nWithFsFile(files...)
	if err != nil {
		global.LOG.Error("initI18n Fail", zap.Error(err), zap.Any("config", global.CONF))
		return err
	}
	global.I18N = i18n
	global.Render = rest.NewI18nRender(i18n)
	global.OK = global.Render.OK
	global.OkWithMsg = global.Render.OkWithMsg
	global.FAIL = global.Render.FAIL
	global.FailWithMsg = global.Render.FailWithMsg
	global.JSON = global.Render.JSON
	global.FailWithError = global.Render.FailWithError
	return nil
}
