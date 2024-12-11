package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/lib/zhttp"
	"go.uber.org/zap"
)

func Request(ctx context.Context, method, url string,
	body io.Reader, obj any, after func(raw []byte, obj any) error) error {
	return zhttp.R.
		Client(global.HttpClient).
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("Cookie", global.GeekCookies)
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Referer", url)
			r.Header.Set("Origin", "https://time.geekbang.org")
		}).
		After(func(r *http.Response) error {
			if zhttp.IsHTTPSuccessStatus(r.StatusCode) {
				switch r.Header.Get("Content-Type") {
				case "application/json",
					"application/json; charset=utf-8":
					raw, err := io.ReadAll(r.Body)
					if err != nil {
						return err
					}
					r.Body = io.NopCloser(bytes.NewReader(raw))
					if err = json.Unmarshal(raw, obj); err != nil {
						global.LOG.Error("Request", zap.String("url", url),
							zap.String("raw", string(raw)),
							zap.Error(err))
						return err
					}
					if after != nil {
						return after(raw, obj)
					}
					return nil
				}
			}
			if zhttp.IsHTTPStatusSleep(r.StatusCode) {
				time.Sleep(time.Second * 10)
			}
			if zhttp.IsHTTPStatusRetryable(r.StatusCode) {
				return errors.New("http status: " + r.Status)
			}
			return zhttp.BreakRetryError(errors.New("http status: " + r.Status))
		}).
		DoWithRetry(ctx, method, url, body)
}
