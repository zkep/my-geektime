package task

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/zkep/mygeektime/internal/types/task"
	"sync"
	"time"

	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
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
		global.LOG.Debug("task handler running", zap.Time("time", t))
		return nil
	}
	defer lock.Delete(keyLock)
	timeCtx, timeCancel := context.WithTimeout(ctx, time.Minute*15)
	defer timeCancel()
	hasMore, page, psize := true, 1, 5
	for hasMore {
		var ls []*model.Task
		t1 := time.Now().AddDate(0, 0, -1).Unix()
		if err := global.DB.Model(&model.Task{}).
			Where("created_at >= ?", t1).
			Where("status = ?", service.TASK_STATUS_PENDING).
			Order("id ASC").
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
		for idx := range ls {
			x := ls[idx]
			if err := worker(timeCtx, x); err != nil {
				global.LOG.Error("task handler worker", zap.Error(err), zap.String("taskid", x.TaskId))
			}
		}
	}
	global.LOG.Debug("task handler End", zap.Time("time", time.Now()))
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
			global.LOG.Error("task handler Count",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
		status := service.TASK_STATUS_FINISHED
		if count > 0 {
			global.LOG.Info("task worker sub task",
				zap.Int64("pending", count),
				zap.String("taskId", x.TaskId),
			)
			status = service.TASK_STATUS_PENDING
		}
		var statistics task.TaskStatistics
		if err := json.Unmarshal(x.Statistics, &statistics); err != nil {
			global.LOG.Error("task worker Unmarshal",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
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
				global.LOG.Error("task handler Count",
					zap.Error(err),
					zap.String("taskId", x.TaskId),
				)
			}
			statistics.Items[item] = int(itemCount)
		}
		raw, _ := json.Marshal(statistics)
		m := map[string]any{
			"status":     status,
			"statistics": raw,
			"updated_at": time.Now().Unix(),
		}
		if err := global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error; err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
	case service.TASK_TYPE_ARTICLE:
		m := map[string]any{
			"status":     service.TASK_STATUS_RUNNING,
			"updated_at": time.Now().Unix(),
		}
		if err := global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error; err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err),
				zap.String("taskId", x.TaskId),
			)
			return err
		}
		message := bytes.NewBuffer(nil)
		status := service.TASK_STATUS_FINISHED
		err := service.Download(ctx, x)
		if err != nil {
			global.LOG.Error("task worker download",
				zap.Error(err), zap.String("taskId", x.TaskId))
			status = service.TASK_STATUS_ERROR
			message.WriteString(err.Error())
		} else {
			message.Write(x.Message)
		}
		m = map[string]any{
			"status":     status,
			"updated_at": time.Now().Unix(),
			"message":    message.Bytes(),
		}
		err = global.DB.Model(&model.Task{Id: x.Id}).UpdateColumns(m).Error
		if err != nil {
			global.LOG.Error("task worker UpdateColumns",
				zap.Error(err), zap.String("taskId", x.TaskId),
			)
			return err
		}
	}
	return nil
}
