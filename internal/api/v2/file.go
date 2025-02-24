package v2

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/storage"
	"go.uber.org/zap"
)

type File struct{}

func NewFile() *File {
	return &File{}
}

func (f *File) Proxy(c *gin.Context) {
	uri, ok := c.GetQuery("url")
	if !ok {
		c.DataFromReader(404, 0, "", nil, nil)
		return
	}
	hash := md5.New()
	hash.Write([]byte(uri))
	hashStr := hex.EncodeToString(hash.Sum(nil))
	cacheKey := filepath.Join(global.CONF.Site.Proxy.CachePrefix, hashStr)
	if global.CONF.Site.Proxy.Cache {
		global.LOG.Info("file.proxy.Get",
			zap.String("cacheKey", cacheKey),
			zap.String("url", uri),
			zap.String("contentType", storage.TypeByExtension(uri)),
		)
		if fi, stat, err := global.Storage.Get(cacheKey); err != nil {
			if !strings.Contains(err.Error(), "no such file or directory") {
				global.LOG.Error("file.proxy.Get", zap.Error(err), zap.String("cacheKey", cacheKey))
				c.DataFromReader(404, 0, "", nil, nil)
				return
			}
		} else {
			c.DataFromReader(200, stat.Size(), storage.TypeByExtension(uri), fi, nil)
			return
		}
	}
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		global.LOG.Error("file.proxy.NewRequest", zap.String("cacheKey", cacheKey), zap.Error(err))
		c.DataFromReader(404, 0, "", nil, nil)
		return
	}
	header, ok := c.GetQuery("header")
	if ok && len(header) > 0 {
		for _, v := range strings.Split(header, ",") {
			headerPair := strings.Split(v, ":")
			if len(headerPair) != 2 {
				continue
			}
			key := strings.TrimSpace(headerPair[0])
			value := strings.TrimSpace(headerPair[1])
			request.Header.Add(key, value)
		}
	}

	request.Header.Set("Referer", uri)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		global.LOG.Error("file.proxy.Do", zap.String("cacheKey", cacheKey), zap.Error(err))
		c.DataFromReader(404, 0, "", nil, nil)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		global.LOG.Error("file.proxy.StatusCode",
			zap.String("cacheKey", cacheKey), zap.Int("statusCode", resp.StatusCode))
		c.DataFromReader(resp.StatusCode, 0, "", nil, nil)
		return
	}
	headers := func(h http.Header) map[string]string {
		m := make(map[string]string)
		for k := range h {
			if k == "Content-Type" {
				continue
			}
			m[k] = h.Get(k)
		}
		return m
	}
	contentType := resp.Header.Get("Content-Type")
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LOG.Error("file.proxy.ReadAll", zap.Error(err), zap.String("cacheKey", cacheKey))
		c.DataFromReader(404, 0, "", nil, nil)
		return
	}
	resp.Body = io.NopCloser(bytes.NewReader(bs))
	if global.CONF.Site.Proxy.Cache {
		go func() {
			if _, err = global.Storage.Put(cacheKey, io.NopCloser(bytes.NewReader(bs))); err != nil {
				global.LOG.Error("file.proxy.Put", zap.Error(err), zap.String("cacheKey", cacheKey))
			}
		}()
	}
	c.DataFromReader(resp.StatusCode, resp.ContentLength, contentType, resp.Body, headers(resp.Header))
}
