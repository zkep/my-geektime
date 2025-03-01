package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/lib/zhttp"
	"go.uber.org/zap"
)

const (
	authURL    = "https://account.geekbang.org/serv/v1/user/auth"
	refererURL = "https://time.geekbang.org/dashboard/usercenter"
)

var (
	ErrorGeekAccountNotLogin = errors.New("geek account not login")
)

func SaveCookie(cookies string, identity string, auth *geek.AuthResponse) func(r *http.Response) error {
	return func(r *http.Response) error {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(raw, auth); err != nil {
			global.LOG.Error("SaveCookie", zap.Error(err), zap.String("raw", string(raw)))
			return err
		}
		if auth.Code != 0 {
			global.LOG.Error("SaveCookie", zap.String("raw", string(raw)))
			return ErrorGeekAccountNotLogin
		}
		var authData geek.GeekUser
		if err = json.Unmarshal(auth.Data, &authData); err != nil {
			global.LOG.Error("SaveCookie", zap.Error(err), zap.String("raw", string(raw)))
			return ErrorGeekAccountNotLogin
		}
		user := model.User{
			Uid:         identity,
			NickName:    authData.Nick,
			Avatar:      authData.Avatar,
			AccessToken: cookies,
		}
		if err = global.DB.Where(model.User{Uid: identity}).
			Assign(model.User{
				Avatar:      authData.Avatar,
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

	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("Referer", refererURL)
		r.Header.Set("Cookie", cookies)
		r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "https://time.geekbang.com")
	}
	err := zhttp.NewRequest().Client(global.HttpClient).Before(before).
		After(after).DoWithRetry(context.Background(), http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	return nil
}
