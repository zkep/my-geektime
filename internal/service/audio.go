package service

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/libs/zhttp"
)

func Audio(ctx context.Context, _ *model.Task, dowloadURL, dir, fileName string) (string, error) {
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute*5)
	defer retryCancel()
	destName := path.Join(dir, fmt.Sprintf("%s.mp3", fileName))

	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
		r.Header.Set("Referer", r.URL.String())
		r.Header.Set("Origin", "https://time.geekbang.org")
	}

	after := func(r *http.Response) error {
		if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
			stat, err := global.Storage.Put(destName, r.Body)
			if err != nil {
				return err
			} else if stat.Size() <= 0 {
				return fmt.Errorf("audio empty: [%s]", destName)
			}
			destName = global.Storage.GetKey(destName, true)
			return nil
		}
		if zhttp.IsHTTPStatusSleep(r.StatusCode) {
			time.Sleep(time.Second * 10)
		}
		if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
			return fmt.Errorf("http status: %s", r.Status)
		}
		return zhttp.BreakRetryError(fmt.Errorf("http status: %s", r.Status))
	}

	err := zhttp.NewRequest().Before(before).After(after).
		DoWithRetry(retryCtx, http.MethodGet, dowloadURL, nil)
	if err != nil {
		return "", err
	}
	return destName, nil
}
