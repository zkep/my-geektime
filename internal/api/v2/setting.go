package v2

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/types/setting"
	"gopkg.in/yaml.v3"
)

type Setting struct{}

func NewSetting() *Setting {
	return &Setting{}
}

func (s *Setting) Update(c *gin.Context) {
	var req setting.SettingUpdate
	if err := c.BindJSON(&req); err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	global.CONF.Storage.Host = req.StorageHost
	global.CONF.Site.Download = req.SiteDownload
	global.CONF.Site.Cache = req.SiteCache
	global.CONF.Site.Play.ProxyUrl = req.SitePlayUrls
	global.CONF.Site.Proxy.ProxyUrl = req.SiteProxyURL
	global.CONF.Site.Proxy.Urls = req.SiteProxyUrls
	raw, err := yaml.Marshal(global.CONF)
	if err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	customConfPath := global.CustomConfigFile
	if len(global.CONFPath) > 0 {
		customConfPath = filepath.Join(filepath.Dir(global.CONFPath), global.CustomConfigFile)
	}
	if err = os.WriteFile(customConfPath, raw, os.ModePerm); err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	global.OK(c, nil)
}

func (s *Setting) Query(c *gin.Context) {
	resp := setting.QueryResponse{
		Storage: global.CONF.Storage,
		Site:    global.CONF.Site,
	}
	customConfPath := global.CustomConfigFile
	if len(global.CONFPath) > 0 {
		customConfPath = filepath.Join(filepath.Dir(global.CONFPath), global.CustomConfigFile)
	}
	if stat, err := os.Stat(customConfPath); err == nil && stat.Size() > 0 {
		var cfg config.Config
		raw, err1 := os.ReadFile(customConfPath)
		if err1 != nil {
			global.FAIL(c, "fail.msg", err1)
			return
		}
		if err = yaml.Unmarshal(raw, &cfg); err != nil {
			global.FAIL(c, "fail.msg", err)
			return
		}
		resp.Site = cfg.Site
		resp.Storage = cfg.Storage
	}
	global.OK(c, resp)
}
