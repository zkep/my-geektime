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
	"github.com/zkep/my-geektime/internal/types/task"
	"github.com/zkep/my-geektime/internal/types/user"
	"github.com/zkep/my-geektime/libs/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type DataFlags struct {
	Config   string  `name:"config" description:"Path to config file"`
	Ids      []int32 `name:"id" description:"1: 体系课，4:公开课" default:"1"`
	Cookies  string  `name:"cookies" description:"geektime cookies string"`
	Download bool    `name:"download" description:"download geektime source" default:"false"`
}

func (app *App) Data(f *DataFlags) error {
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
	global.CONF.Site.Download = f.Download
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
	tagRaw, err := app.assets.ReadFile("web/pages/tags.json")
	if err != nil {
		return err
	}
	var tagData sys_dict.TagData
	if err = json.Unmarshal(tagRaw, &tagData); err != nil {
		return err
	}
	for _, id := range f.Ids {
		typ, ok := sys_dict.ProductTypes[id]
		if !ok {
			fmt.Printf("not found product id [%d]", id)
			continue
		}
		for _, form := range sys_dict.ProductForms {
			for _, tag := range tagData.Data {
				for _, opt := range tag.Options {
					if err = app.iterators(typ, form, opt, tag, id, accessToken); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (app *App) iterators(typ, form, opt sys_dict.Option, tag sys_dict.Tag, id int32, accessToken string) error {
	prev, psize, hasNext, total := 0, 20, true, 0
	fmt.Printf(
		"download start [%s/%s/%s/%s] \n",
		typ.Label, form.Label, tag.Label, opt.Label,
	)
	for hasNext {
		req := geek.PvipProductRequest{
			TagIds:       []int32{opt.Value},
			ProductType:  id,
			ProductForm:  form.Value,
			Sort:         8,
			Size:         psize,
			Prev:         prev,
			WithArticles: true,
		}
		resp, err := service.GetPvipProduct(app.ctx, accessToken, req)
		if err != nil {
			return err
		}
		total += len(resp.Data.Products)
		if len(resp.Data.Products) < psize {
			fmt.Printf(
				"download end [%s/%s/%s/%s] total: %d , pageTotal: %d \n",
				typ.Label, form.Label, tag.Label, opt.Label, total, resp.Data.Page.Total,
			)
			hasNext = false
		}
		prev++
		for _, product := range resp.Data.Products {
			articles, err1 := service.GetArticles(app.ctx, accessToken, geek.ArticlesListRequest{
				Cid:   fmt.Sprintf("%d", product.ID),
				Order: "earliest",
				Prev:  0,
				Size:  500,
			})
			if err1 != nil {
				return err1
			}
			jobId := utils.HalfUUID()
			itemRaw, _ := json.Marshal(product)
			job := &model.Task{
				TaskId:     jobId,
				TaskName:   product.Title,
				TaskType:   service.TASK_TYPE_PRODUCT,
				OtherId:    fmt.Sprintf("%d", product.ID),
				Cover:      product.Cover.Square,
				Raw:        itemRaw,
				OtherType:  typ.Value,
				OtherForm:  form.Value,
				OtherGroup: tag.Value,
				OtherTag:   opt.Value,
				Status:     service.TASK_STATUS_PENDING,
			}
			tasks := make([]*model.Task, 0, len(articles.Data.List))
			for _, article := range articles.Data.List {
				info, er := service.GetArticleInfo(app.ctx,
					accessToken, geek.ArticlesInfoRequest{Id: article.ID})
				if er != nil {
					return er
				}
				var m geek.ArticleInfoRaw
				if err = json.Unmarshal(info.Raw, &m); err != nil {
					return err
				}
				raw := m.Data
				otherId := fmt.Sprintf("%d", info.Data.Info.ID)
				taskName := info.Data.Info.Title
				cover := info.Data.Info.Cover.Default
				item := model.Task{
					TaskPid:    jobId,
					TaskId:     utils.HalfUUID(),
					OtherId:    otherId,
					TaskName:   taskName,
					TaskType:   service.TASK_TYPE_ARTICLE,
					Cover:      cover,
					Raw:        raw,
					OtherType:  typ.Value,
					OtherForm:  form.Value,
					OtherGroup: tag.Value,
					OtherTag:   opt.Value,
					Status:     service.TASK_STATUS_PENDING,
				}
				tasks = append(tasks, &item)
			}
			statistics := task.TaskStatistics{
				Count: len(tasks),
				Items: map[int]int{
					service.TASK_STATUS_PENDING:  len(tasks),
					service.TASK_STATUS_RUNNING:  0,
					service.TASK_STATUS_FINISHED: 0,
					service.TASK_STATUS_ERROR:    0,
				},
			}
			job.Statistics, _ = json.Marshal(statistics)
			err = global.DB.Transaction(func(tx *gorm.DB) error {
				if err = tx.Where(&model.Task{OtherId: job.OtherId}).
					Assign(job).FirstOrCreate(job).Error; err != nil {
					return err
				}
				for _, x := range tasks {
					if err = tx.Where(&model.Task{OtherId: x.OtherId}).
						Assign(x).FirstOrCreate(x).Error; err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
