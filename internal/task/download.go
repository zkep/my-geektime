package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
	"github.com/zkep/mygeektime/internal/types/user"
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
		if err := iterators(ctx, true); err != nil {
			global.LOG.Error("task handler iterators", zap.Error(err), zap.Bool("loaded", loaded))
		}
		return nil
	}
	defer lock.Delete(keyLock)
	if err := iterators(ctx, false); err != nil {
		global.LOG.Error("task handler iterators", zap.Error(err), zap.Bool("loaded", loaded))
	}
	global.LOG.Debug("task handler End", zap.Time("time", time.Now()))
	return nil
}

func iterators(ctx context.Context, loaded bool) error {
	timeCtx, timeCancel := context.WithTimeout(ctx, time.Hour)
	defer timeCancel()
	hasMore, page, psize := true, 1, 6
	orderTasks, batchTasks := make([]*model.Task, 0, psize), make([]*model.Task, 0, psize)
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
		orderTasks = orderTasks[:0]
		batchTasks = batchTasks[:0]
		for _, value := range ls {
			if len(value.RewriteHls) == 0 {
				orderTasks = append(orderTasks, value)
			} else {
				batchTasks = append(batchTasks, value)
			}
		}

		batch := global.GPool.NewBatch()
		for _, value := range batchTasks {
			x := value
			batch.Queue(func(pctx context.Context) (any, error) {
				err := worker(pctx, x)
				if err != nil {
					global.LOG.Error("task handler worker", zap.Error(err), zap.String("taskid", x.TaskId))
				}
				return nil, err
			})
		}
		for _, value := range orderTasks {
			x := value
			batch.Queue(func(pctx context.Context) (any, error) {
				err := worker(pctx, x)
				if err != nil {
					global.LOG.Error("task handler worker", zap.Error(err), zap.String("taskid", x.TaskId))
				}
				return nil, err
			})
		}
		if _, err := batch.Wait(timeCtx); err != nil {
			global.LOG.Error("task handler wait", zap.Error(err))
			return err
		}
	}
	return nil
}

func worker(ctx context.Context, x *model.Task) error {
	switch x.TaskType {
	case service.TASK_TYPE_PRODUCT:
		return doProduct(ctx, x)
	case service.TASK_TYPE_ARTICLE:
		return doArticle(ctx, x)
	default:
	}
	return nil
}

func doProduct(_ context.Context, x *model.Task) error {
	var statistics task.TaskStatistics
	if err := json.Unmarshal(x.Statistics, &statistics); err != nil {
		global.LOG.Error("worker Unmarshal", zap.Error(err), zap.String("taskId", x.TaskId))
	}
	if statistics.Items == nil {
		statistics.Items = make(map[int]int, 5)
	}
	pendingCount, runingCount, errorCount := int64(0), int64(0), int64(0)
	for _, item := range service.ALLStatus {
		var itemCount int64
		if err := global.DB.Model(&model.Task{}).
			Where("task_pid = ?", x.TaskId).
			Where("status = ?", item).
			Count(&itemCount).Error; err != nil {
			global.LOG.Error("worker count", zap.Error(err), zap.String("taskId", x.TaskId))
		}
		switch item {
		case service.TASK_STATUS_PENDING:
			pendingCount = itemCount
		case service.TASK_STATUS_RUNNING:
			runingCount = itemCount
		case service.TASK_STATUS_ERROR:
			errorCount = itemCount
		}
		statistics.Items[item] = int(itemCount)
	}
	status := service.TASK_STATUS_FINISHED
	if pendingCount > 0 {
		status = service.TASK_STATUS_PENDING
	}
	if runingCount > 0 {
		status = service.TASK_STATUS_PENDING
	}
	if errorCount > 0 {
		status = service.TASK_STATUS_ERROR
	}
	raw, _ := json.Marshal(statistics)
	m := model.Task{
		Id:         x.Id,
		Status:     int32(status),
		Statistics: raw,
		UpdatedAt:  time.Now().Unix(),
	}
	if x.Bstatus > 0 {
		if status == service.TASK_STATUS_FINISHED {
			var product geek.ProductBase
			if len(x.Raw) > 0 {
				_ = json.Unmarshal(x.Raw, &product)
			}
			dir := path.Join(x.TaskId, service.VerifyFileName(product.Title))
			dirURL := global.Storage.GetKey(dir, false)
			message := task.TaskMessage{Object: dirURL}
			m.Message, _ = json.Marshal(message)
		}
	}
	if err := global.DB.Model(&model.Task{}).Where(&model.Task{Id: x.Id}).Updates(&m).Error; err != nil {
		global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
		return err
	}
	return nil
}

func doArticle(ctx context.Context, x *model.Task) error {
	m := model.Task{
		Id:        x.Id,
		Status:    service.TASK_STATUS_RUNNING,
		UpdatedAt: time.Now().Unix(),
	}
	var data geek.ArticleData
	if len(x.RewriteHls) > 0 {
		if len(x.Ciphertext) == 0 && bytes.Contains(x.RewriteHls, []byte("{host}/v2/task/kms")) {
			x.RewriteHls = []byte(``)
		}
		if !bytes.Contains(x.RewriteHls, []byte("#EXTM3U")) {
			x.RewriteHls = []byte(``)
		}
	}
	if len(x.RewriteHls) == 0 {
		aid, err := strconv.ParseInt(x.OtherId, 10, 64)
		if err != nil {
			global.LOG.Error("worker ParseInt", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		var u model.User
		if err = global.DB.Where(&model.User{RoleId: user.AdminRoleId}).First(&u).Error; err != nil {
			global.LOG.Error("worker User", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		if u.AccessToken == "" {
			return errors.New("no access token, please refresh geektime cookie")
		}
		var auth geek.AuthResponse
		if err = service.Authority(u.AccessToken, service.SaveCookie(u.AccessToken, u.Uid, &auth)); err != nil {
			global.LOG.Error("worker Authority", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		article, err1 := service.GetArticleInfo(ctx, u.Uid, u.AccessToken, geek.ArticlesInfoRequest{Id: aid})
		if err1 != nil {
			global.LOG.Error("worker GetArticleInfo", zap.Error(err1), zap.String("taskId", x.TaskId))
			return err1
		}
		if err = service.ArticleAllComment(ctx, u.Uid, u.AccessToken, aid); err != nil {
			global.LOG.Error("worker ArticleAllComment", zap.Error(err), zap.String("taskId", x.TaskId))
		}
		data = article.Data
		var info geek.ArticleInfoRaw
		if err = json.Unmarshal(article.Raw, &info); err != nil {
			global.LOG.Error("worker Unmarshal", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
		m.Raw = info.Data
		x.Raw = info.Data
	} else {
		if err := json.Unmarshal(x.Raw, &data); err != nil {
			global.LOG.Error("worker Unmarshal", zap.Error(err), zap.String("taskId", x.TaskId))
			return err
		}
	}
	if err := global.DB.Where(&model.Task{Id: x.Id}).Updates(m).Error; err != nil {
		global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
		return err
	}
	status := service.TASK_STATUS_FINISHED
	if err := service.Download(ctx, x, data); err != nil {
		global.LOG.Error("worker download", zap.Error(err), zap.String("taskId", x.TaskId))
		status = service.TASK_STATUS_ERROR
		message := task.TaskMessage{Text: err.Error()}
		x.Message, _ = json.Marshal(message)
	}
	m.Ciphertext = x.Ciphertext
	m.RewriteHls = x.RewriteHls
	m.Message = x.Message
	m.Status = int32(status)
	m.UpdatedAt = time.Now().Unix()
	if err := global.DB.Where(&model.Task{Id: x.Id}).Updates(&m).Error; err != nil {
		global.LOG.Error("worker Updates", zap.Error(err), zap.String("taskId", x.TaskId))
		return err
	}
	return nil
}
