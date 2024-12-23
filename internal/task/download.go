package task

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
	"github.com/zkep/mygeektime/internal/types/user"
	"github.com/zkep/mygeektime/lib/pool"
	"go.uber.org/zap"
)

var (
	keyLock = "_key_lock_"
	lock    = &sync.Map{}
)

func TaskHandler(ctx context.Context, t time.Time) error {
	global.LOG.Debug("task handler Start", zap.Time("time", t))
	_, loaded := lock.LoadOrStore(keyLock, t)
	if loaded {
		if err := iterators(ctx, global.GPool, true); err != nil {
			global.LOG.Error("task handler iterators", zap.Error(err), zap.Bool("loaded", loaded))
		}
		return nil
	}
	defer lock.Delete(keyLock)
	if err := iterators(ctx, global.GPool, false); err != nil {
		global.LOG.Error("task handler iterators", zap.Error(err), zap.Bool("loaded", loaded))
	}
	global.LOG.Debug("task handler End", zap.Time("time", time.Now()))
	return nil
}

func iterators(ctx context.Context, gpool *pool.GPool, loaded bool) error {
	timeCtx, timeCancel := context.WithTimeout(ctx, time.Hour)
	defer timeCancel()
	hasMore, page, psize := true, 1, 5
	batch := gpool.NewBatch()
	for hasMore {
		var ls []*model.Task
		tx := global.DB.Model(&model.Task{})
		if loaded {
			tx = tx.Where("task_pid = ?", "").Where("status <= ?", service.TASK_STATUS_PENDING)
		} else {
			tx = tx.Where("status = ?", service.TASK_STATUS_PENDING)
		}
		tx = tx.Where("deleted_at = ?", 0)
		if err := tx.Order("id ASC").
			Offset((page - 1) * psize).
			Limit(psize + 1).
			Find(&ls).Error; err != nil {
			global.LOG.Error("task handler find", zap.Error(err))
			return err
		}
		if len(ls) <= psize {
			hasMore = false
		} else {
			ls = ls[:psize]
		}
		page++
		for _, value := range ls {
			x := value
			batch.Queue(func(pctx context.Context) (any, error) {
				if err := worker(pctx, x); err != nil {
					global.LOG.Error("task handler worker", zap.Error(err), zap.String("taskid", x.TaskId))
					return nil, err
				}
				return nil, nil
			})
		}
	}
	if _, err := batch.Wait(timeCtx); err != nil {
		global.LOG.Error("task handler wait", zap.Error(err))
		return err
	}
	return nil
}

func worker(ctx context.Context, x *model.Task) error {
	switch x.TaskType {
	case service.TASK_TYPE_PRODUCT:
		var count int64
		if err := global.DB.Model(&model.Task{}).
			Where("task_pid = ?", x.TaskId).
			Where("status <= ?", service.TASK_STATUS_RUNNING).
			Count(&count).Error; err != nil {
			global.LOG.Error("worker", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		status := service.TASK_STATUS_FINISHED
		if count > 0 {
			global.LOG.Info("worker sub task",
				zap.Int64("pending", count), zap.String("taskId", x.TaskId))
			status = service.TASK_STATUS_PENDING
		}
		var statistics task.TaskStatistics
		if err := json.Unmarshal(x.Statistics, &statistics); err != nil {
			global.LOG.Error("worker Unmarshal", zap.Error(err), zap.String("taskId", x.TaskId))
		}
		if statistics.Items == nil {
			statistics.Items = make(map[int]int, 5)
		}
		for _, item := range service.ALLStatus {
			var itemCount int64
			if err := global.DB.Model(&model.Task{}).
				Where("task_pid = ?", x.TaskId).
				Where("status = ?", item).
				Count(&itemCount).Error; err != nil {
				global.LOG.Error("worker count", zap.Error(err), zap.String("taskId", x.TaskId))
			}
			statistics.Items[item] = int(itemCount)
		}
		raw, _ := json.Marshal(statistics)
		m := model.Task{
			Id:         x.Id,
			Status:     int32(status),
			Statistics: raw,
			UpdatedAt:  time.Now().Unix(),
		}
		if err := global.DB.Where(&model.Task{Id: x.Id}).Updates(&m).Error; err != nil {
			global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
	case service.TASK_TYPE_ARTICLE:
		m := model.Task{
			Id:        x.Id,
			Status:    service.TASK_STATUS_RUNNING,
			UpdatedAt: time.Now().Unix(),
		}
		aid, err := strconv.ParseInt(x.OtherId, 10, 64)
		if err != nil {
			return err
		}
		var u model.User
		if err = global.DB.Where(&model.User{RoleId: user.AdminRoleId}).First(&u).Error; err != nil {
			return err
		}
		if u.AccessToken == "" {
			return errors.New("no access token, please refresh geektime cookie")
		}
		status := service.TASK_STATUS_FINISHED
		if article, err1 := service.GetArticleInfo(ctx, u.Uid,
			u.AccessToken, geek.ArticlesInfoRequest{Id: aid}); err1 != nil {
			return err1
		} else {
			var info geek.ArticleInfoRaw
			if err = json.Unmarshal(article.Raw, &info); err != nil {
				return err
			}
			m.Raw = info.Data
		}
		if global.CONF.Site.Download {
			if err = global.DB.Where(&model.Task{Id: x.Id}).Updates(m).Error; err != nil {
				global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
				return err
			}
			if err = service.Download(ctx, x); err != nil {
				global.LOG.Error("worker download", zap.Error(err), zap.String("taskId", x.TaskId))
				status = service.TASK_STATUS_ERROR
				message := task.TaskMessage{Text: err.Error()}
				x.Message, _ = json.Marshal(message)
			}
		}
		m.Message = x.Message
		m.Status = int32(status)
		m.UpdatedAt = time.Now().Unix()
		err = global.DB.Where(&model.Task{Id: x.Id}).Updates(&m).Error
		if err != nil {
			global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
	}
	return nil
}
