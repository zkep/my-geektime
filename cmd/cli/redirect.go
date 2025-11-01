package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/initialize"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/internal/types/user"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type RedirectFlags struct {
	Config  string   `name:"config" description:"Path to config file"`
	Types   []string `name:"types" description:"1: 体系课，4:公开课，5:线下大会，6:社区课，d:每日一课，q:大厂案例"`
	Cookies string   `name:"cookies" description:"geektime cookies string"`
	Detail  bool     `name:"detail" description:"detail url"`
}

func (app *App) Redirect(f *RedirectFlags) error {
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

	fmt.Printf("database exists product [%d]\n\n", len(dataTasksMap))

	types := make([]int32, 0, len(f.Types))
	for _, typ := range f.Types {
		otherType, ok := sys_dict.ProductTypes[typ]
		if !ok {
			fmt.Printf("not found product typ [%s]", typ)
			continue
		}
		types = append(types, otherType.Value)
	}

	page, size, hasNext := 1, 20, true
	for hasNext {
		var ls []*model.Task
		tx := global.DB.Model(&model.Task{}).
			Select(
				[]string{
					"id", "task_id", "other_id", "other_type", "task_type",
					"other_tag", "other_form", "raw",
				},
			)
		if len(types) > 0 {
			tx = tx.Where("other_type IN ?", types)
		}
		if !f.Detail {
			tx = tx.Where("task_pid=?", "")
		}
		if err = tx.Offset((page - 1) * size).
			Limit(size + 1).Find(&ls).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		page++
		if len(ls) > size {
			ls = ls[:size]
		} else {
			hasNext = false
		}
		for _, x := range ls {
			switch x.TaskType {
			case service.TASK_TYPE_PRODUCT:
				var product geek.ProductBase
				if len(x.Raw) > 0 {
					_ = json.Unmarshal(x.Raw, &product)
				}
				redirect := sys_dict.ProductURLWithType(product.Type, product.ID)
				fmt.Printf("product task_id: %s, type: %s, redirect: %s\n", x.TaskId, product.Type, redirect)
			case service.TASK_TYPE_ARTICLE:
				var articleInfo geek.ArticleData
				if len(x.Raw) > 0 {
					_ = json.Unmarshal(x.Raw, &articleInfo)
				}
				redirect := sys_dict.ProductDetailURLWithType(articleInfo.Product.Type, articleInfo.Product.ID, articleInfo.Info.ID)
				fmt.Printf("article task_id: %s, type: %s, articleType: %d, redirect: %s\n",
					x.TaskId, articleInfo.Product.Type, articleInfo.Info.Type, redirect)
			}
			dataTasksMap[x.OtherId] = &DataTask{
				TaskId:     x.TaskId,
				OtherType:  x.OtherType,
				OtherTag:   x.OtherTag,
				OtherForm:  x.OtherForm,
				OtherGroup: x.OtherGroup,
			}
		}
	}

	return nil
}
