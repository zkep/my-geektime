package cli

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/initialize"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/internal/types/user"
	"github.com/zkep/my-geektime/libs/zhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type LableResponse struct {
	Error []any `json:"error,omitempty"`
	Extra []any `json:"extra,omitempty"`
	Data  Data  `json:"data,omitempty"`
	Code  int   `json:"code,omitempty"`
}

type Data struct {
	Nav    []Nav  `json:"nav,omitempty"`
	Labels []Item `json:"labels,omitempty"`
}

type Nav struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

type Item struct {
	DisplayType int    `json:"display_type,omitempty"`
	Lid         int32  `json:"lid,omitempty"`
	Count       int    `json:"count,omitempty"`
	Sort        int    `json:"sort,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Name        string `json:"name,omitempty"`
	EsIcon      string `json:"es_icon,omitempty"`
	Pid         int32  `json:"pid,omitempty"`
}

type Node struct {
	Item
	Children []*Node `json:"children,omitempty"`
}

const (
	labelURL = "https://time.geekbang.org/serv/v1/column/labels"
)

type LabelFlags struct {
	Config  string `name:"config" description:"Path to config file"`
	Cookies string `name:"cookies" description:"geektime cookies string"`
}

func (app *App) Label(f *LabelFlags) error {
	var (
		cfg         config.Config
		accessToken string
		configRaw   []byte
		err         error
	)
	if f.Config == "" {
		configRaw, err = app.assets.ReadFile("config.yml")
	} else {
		configRaw, err = os.ReadFile(f.Config)
	}
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(configRaw, &cfg); err != nil {
		return err
	}
	global.CONF = &cfg
	global.ASSETS = app.assets
	if err = initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err = initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err = initialize.Storage(app.ctx); err != nil {
		return err
	}
	if len(f.Cookies) > 0 {
		accessToken = f.Cookies
		if cookies := os.Getenv("cookies"); len(cookies) > 0 {
			accessToken = cookies
		}
	} else {
		var u model.User
		if err = global.DB.
			Where(&model.User{RoleId: user.AdminRoleId}).
			First(&u).Error; err != nil {
			return err
		}
		accessToken = u.AccessToken
	}
	if accessToken == "" {
		return errors.New("no access token")
	}
	after := func(r *http.Response) error {
		var auth geek.AuthResponse
		authData, err1 := service.GetGeekUser(r, &auth)
		if err1 != nil {
			global.LOG.Error("GetGeekUser", zap.Error(err1))
			return err1
		}
		if authData.UID <= 0 {
			return fmt.Errorf("no user")
		}
		return nil
	}
	if err = service.Authority(accessToken, after); err != nil {
		return err
	}
	before := func(r *http.Request) {
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("Referer", labelURL)
		r.Header.Set("Cookie", accessToken)
		r.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
		r.Header.Set("User-Agent", zhttp.RandomUserAgent())
		r.Header.Set("Accept", "application/json, text/plain, */*")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "https://time.geekbang.com")
	}
	err = zhttp.NewRequest().Before(before).
		After(func(r *http.Response) error {
			raw, err1 := io.ReadAll(r.Body)
			if err1 != nil {
				return err
			}
			var resp LableResponse
			if err = json.Unmarshal(raw, &resp); err != nil {
				return err
			}
			l := list.New()
			node := &Node{}
			for _, x := range resp.Data.Labels {
				x.Name = strings.ReplaceAll(x.Name, "/", "-")
				l.PushBack(&x)
			}
			for e := l.Front(); e != nil; e = e.Next() {
				x, ok := e.Value.(*Item)
				if !ok {
					continue
				}
				child := &Node{Item: *x, Children: make([]*Node, 0, 10)}
				if x.Pid == 0 {
					node.Children = append(node.Children, child)
				} else {
					for k, n := range node.Children {
						if x.Pid == n.Lid {
							n.Children = append(n.Children, child)
							node.Children[k] = n
							break
						}
					}
				}
			}
			sort.Slice(node.Children, func(i, j int) bool {
				return node.Children[i].Lid < node.Children[j].Lid
			})
			tags := make([]sys_dict.Tag, 0, len(node.Children))
			for _, x := range node.Children {
				sort.Slice(x.Children, func(i, j int) bool {
					return x.Children[i].Lid < x.Children[j].Lid
				})
				currTag := sys_dict.Tag{
					Option:  sys_dict.Option{Label: x.Name, Value: x.Lid},
					Options: make([]sys_dict.Option, 0, len(x.Children)),
				}
				for _, child := range x.Children {
					currTag.Options = append(currTag.Options, sys_dict.Option{Label: child.Name, Value: child.Lid})
				}
				tags = append(tags, currTag)
			}
			tagData := sys_dict.TagData{Data: tags}
			raw, _ = json.MarshalIndent(tagData, "", "    ")
			tagFilePath := filepath.Join("web/pages", "tags.json")
			if err = os.WriteFile(tagFilePath, raw, os.ModePerm); err != nil {
				return err
			}
			if err = service.GeektimeCategory(context.Background(), tagData); err != nil {
				return err
			}
			return nil
		}).
		DoWithRetry(
			context.Background(),
			http.MethodPost, labelURL,
			bytes.NewBuffer([]byte(`{"type":0}`)),
		)
	if err != nil {
		return err
	}
	return nil
}
