package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/lib/zhttp"
)

const (
	authURL    = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL = "https://time.geekbang.org/dashboard/usercenter"
)

func SaveCookie(cookies string, identity string, auth *geek.AuthResponse) func(r *http.Response) error {
	return func(r *http.Response) error {
		if err := json.NewDecoder(r.Body).Decode(auth); err != nil {
			return err
		}
		user := model.User{
			Uid:         identity,
			NickName:    auth.Data.Nick,
			Avatar:      auth.Data.Avatar,
			AccessToken: cookies,
		}
		if err := global.DB.Where(model.User{Uid: identity}).
			Assign(model.User{
				Avatar:      auth.Data.Avatar,
				AccessToken: cookies,
			}).
			FirstOrCreate(&user).Error; err != nil {
			return err
		}
		return nil
	}
}

func Authority(cookies string, after func(*http.Response) error) error {
	jar, _ := cookiejar.New(nil)
	global.HttpClient = &http.Client{Jar: jar, Timeout: 5 * time.Minute}
	t := time.Now().UnixMilli()
	authUrl := fmt.Sprintf("%s?t=%d&v_t=%d", authURL, t, t)

	err := zhttp.R.Client(global.HttpClient).
		Before(func(r *http.Request) {
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Referer", refererURL)
			r.Header.Set("Cookie", cookies)
			r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
			r.Header.Set("User-Agent", zhttp.RandomUserAgent())
			r.Header.Set("Accept", "application/json, text/plain, */*")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Origin", "https://time.geekbang.com")
		}).
		After(after).
		DoWithRetry(context.Background(), http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	return nil
}
