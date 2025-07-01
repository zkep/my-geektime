package v2

import (
	"os"

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

const (
	CustomConfigFile = "custom_config.yaml"
)

func (s *Setting) Update(c *gin.Context) {
	var req setting.SettingUpdate
	if err := c.BindJSON(&req); err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	global.CONF.Storage.Host = req.StorageHost
	global.CONF.Site.Download = req.SiteDownload
	global.CONF.Site.Play.ProxyUrl = req.SitePlayUrls
	global.CONF.Site.Proxy.ProxyUrl = req.SiteProxyURL
	global.CONF.Site.Proxy.Urls = req.SiteProxyUrls
	raw, err := yaml.Marshal(global.CONF)
	if err != nil {
		global.FAIL(c, "fail.msg", err)
		return
	}
	if err = os.WriteFile(CustomConfigFile, raw, os.ModePerm); err != nil {
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
	if stat, err := os.Stat(CustomConfigFile); err == nil && stat.Size() > 0 {
		var cfg config.Config
		fi, err1 := os.Open(CustomConfigFile)
		if err1 != nil {
			global.FAIL(c, "fail.msg", err1)
			return
		}
		defer func() { _ = fi.Close() }()
		if err = yaml.NewDecoder(fi).Decode(&cfg); err != nil {
			global.FAIL(c, "fail.msg", err)
			return
		}
		resp.Site = cfg.Site
		resp.Storage = cfg.Storage
	}
	global.OK(c, resp)
}
