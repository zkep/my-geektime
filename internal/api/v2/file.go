package v2

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
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
		c.DataFromReader(404, 0, "", nil, nil)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
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
	c.DataFromReader(resp.StatusCode, resp.ContentLength, contentType, resp.Body, headers(resp.Header))
}
