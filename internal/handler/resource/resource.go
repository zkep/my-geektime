package resource

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/zkep/my-geektime/libs/storage"
)

type Resource struct {
	ctx        context.Context
	url        chan string
	limiter    chan struct{}
	cacheKeyFn func(uri string) string
	matchFn    func(uri string) bool
	storage    storage.Storage
}

func NewResource(
	ctx context.Context, size int,
	cacheKeyFn func(uri string) string,
	matchFn func(uri string) bool,
	storage storage.Storage,
) *Resource {
	if size <= 0 {
		size = 6
	}
	r := &Resource{
		ctx:        ctx,
		limiter:    make(chan struct{}, size),
		url:        make(chan string, 100000),
		cacheKeyFn: cacheKeyFn,
		matchFn:    matchFn,
		storage:    storage,
	}
	r.start()
	return r
}

func (r *Resource) Push(urls ...string) {
	for _, uri := range urls {
		if r.matchFn(uri) {
			r.url <- uri
		}
	}
}

func (r *Resource) start() {
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case v := <-r.url:
				r.limiter <- struct{}{}
				go func(uri string) { r.worker(uri) }(v)
			}
		}
	}()
}

func (r *Resource) worker(uri string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("worker panic:", "err", err, "stack", string(debug.Stack()))
		}
		<-r.limiter
	}()
	if len(uri) == 0 {
		return
	}
	cacheKey := r.cacheKeyFn(uri)
	if stat, _ := r.storage.Stat(cacheKey); stat != nil && stat.Size() > 0 {
		slog.Info("worker cacheKey exists", "cacheKey", cacheKey)
		return
	}
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		slog.Error("worker NewRequest:", "uri", uri, "err", err)
		return
	}
	request.Header.Set("Referer", uri)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("worker Do:", "uri", uri, "err", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		slog.Error("worker StatusCode:", "uri", uri, "StatusCode", resp.StatusCode)
		return
	}
	if _, err = r.storage.Put(cacheKey, io.NopCloser(resp.Body)); err != nil {
		slog.Error("worker Put:", "uri", uri, "err", err)
		return
	}
}

func (r *Resource) Stop() {
	close(r.url)
}
