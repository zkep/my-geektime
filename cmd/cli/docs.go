package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/initialize"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/task"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type DocsFlags struct {
	Config string `name:"config" description:"Path to config file"`
	TaskID string `name:"taskid" description:"task id" default:""`
}

func (app *App) parse(f *DocsFlags, cfg *config.Config) error {
	if f.Config == "" {
		fi, err := app.assets.Open("config.yml")
		if err != nil {
			return err
		}
		defer func() { _ = fi.Close() }()
		if err = yaml.NewDecoder(fi).Decode(cfg); err != nil {
			return err
		}
	} else {
		fi, err := os.Open(f.Config)
		if err != nil {
			return err
		}
		defer func() { _ = fi.Close() }()
		if err = yaml.NewDecoder(fi).Decode(cfg); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) Docs(f *DocsFlags) error {
	var cfg config.Config
	if err := app.parse(f, &cfg); err != nil {
		return err
	}
	global.CONF = &cfg
	if err := initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err := initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err := initialize.Storage(app.ctx); err != nil {
		return err
	}
	if err := initialize.GPool(app.ctx); err != nil {
		return err
	}

	hasMore, page, psize := true, 1, 6
	for hasMore {
		var ls []*model.Task
		tx := global.DB.Model(&model.Task{})
		if len(f.TaskID) > 0 {
			tx = tx.Where("task_id = ?", f.TaskID)
		} else {
			tx = tx.Where("other_form = ?", 1)
			tx = tx.Where("other_type = ?", 1)
		}
		tx = tx.Where("task_pid = ?", "")
		tx = tx.Where("deleted_at = ?", 0)
		if err := tx.Order("id ASC").
			Offset((page - 1) * psize).
			Limit(psize + 1).
			Find(&ls).Error; err != nil {
			global.LOG.Error("Docs find", zap.Error(err))
			return err
		}
		if len(ls) <= psize {
			hasMore = false
		} else {
			ls = ls[:psize]
		}
		page++
		for _, l := range ls {
			var product geek.ProductBase
			if err := json.Unmarshal(l.Raw, &product); err != nil {
				global.LOG.Error("Docs Unmarshal", zap.Error(err))
				continue
			}
			var taskMessage task.TaskMessage
			if len(l.Message) > 0 {
				if err := json.Unmarshal(l.Message, &taskMessage); err != nil {
					global.LOG.Error("Docs Unmarshal", zap.Error(err))
				}
			}
			docURL, err := service.MakeDocsite(app.ctx, l.TaskId, product.Title, product.IntroHTML)
			if err != nil {
				global.LOG.Error("Docs MakeDocsite", zap.Error(err))
				continue
			}
			taskMessage.Doc = docURL
			l.Message, _ = json.Marshal(taskMessage)
			if err = global.DB.Model(&model.Task{}).
				Where(&model.Task{Id: l.Id}).
				UpdateColumn("message", l.Message).Error; err != nil {
				global.LOG.Error("Docs Updates", zap.Error(err), zap.String("taskId", l.TaskId))
				continue
			}
		}
	}
	return nil
}

func (app *App) LocalDoc(f *DocsFlags) error {
	var cfg config.Config
	if err := app.parse(f, &cfg); err != nil {
		return err
	}
	global.CONF = &cfg
	if err := initialize.Gorm(app.ctx); err != nil {
		return err
	}
	if err := initialize.Logger(app.ctx); err != nil {
		return err
	}
	if err := initialize.Storage(app.ctx); err != nil {
		return err
	}
	if err := initialize.GPool(app.ctx); err != nil {
		return err
	}
	var tags []Tag
	if err := json.Unmarshal([]byte(TagJSON), &tags); err != nil {
		return err
	}
	tagMap := make(map[int32]Option, len(tags))
	for _, tag := range tags {
		tagMap[tag.Value] = tag.Option
	}
	workerFn := func(tagValue int32) error {
		hasMore, page, psize := true, 1, 6
		for hasMore {
			var ls []*model.Task
			tx := global.DB.Model(&model.Task{})
			tx = tx.Where("other_group = ?", tagValue)
			if len(f.TaskID) > 0 {
				tx = tx.Where("task_id = ?", f.TaskID)
			}
			tx = tx.Where("other_form = ?", 1)
			tx = tx.Where("other_type = ?", 1)
			tx = tx.Where("task_pid = ?", "")
			tx = tx.Where("deleted_at = ?", 0)
			if err := tx.Order("id ASC").
				Offset((page - 1) * psize).
				Limit(psize + 1).
				Find(&ls).Error; err != nil {
				global.LOG.Error("Docs find", zap.Error(err))
				return err
			}
			if len(ls) <= psize {
				hasMore = false
			} else {
				ls = ls[:psize]
			}
			page++
			for _, l := range ls {
				var product geek.ProductBase
				if err := json.Unmarshal(l.Raw, &product); err != nil {
					global.LOG.Error("Docs Unmarshal", zap.Error(err))
					continue
				}
				group, ok := tagMap[l.OtherGroup]
				if !ok {
					group = Option{Label: "其它"}
				}
				group.Label = service.VerifyFileName(group.Label)
				product.Title = service.VerifyFileName(product.Title)
				err := service.MakeDocsiteLocal(app.ctx, l.TaskId, group.Label, product.Title, product.IntroHTML, 15)
				if err != nil {
					global.LOG.Error("Docs MakeDocsite", zap.Error(err))
					continue
				}
				dir := path.Join(group.Label, product.Title)
				fmt.Printf("\n[%s] docs dir: %s\n", product.Title, global.Storage.GetKey(dir, true))
			}
		}
		return nil
	}
	for _, tag := range tags {
		err := workerFn(tag.Value)
		if err != nil {
			return err
		}
	}
	return nil
}
