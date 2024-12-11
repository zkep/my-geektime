package v2

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
	"github.com/zkep/mygeektime/internal/types/task"
	"gorm.io/gorm"
)

type Task struct{}

func NewTask() *Task {
	return &Task{}
}

func (t *Task) List(c *gin.Context) {
	var req task.TaskListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	if req.PerPage <= 0 || (req.PerPage > 200) {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	ret := task.TaskListResponse{
		Rows: make([]task.Task, 0, req.PerPage),
	}
	var ls []*model.Task
	tx := global.DB.Model(&model.Task{})
	if req.Xstatus > 0 {
		tx = tx.Where("status = ?", req.Xstatus)
	}
	tx = tx.Where("task_pid = ?", req.TaskPid)
	if req.TaskPid != "" {
		tx = tx.Order("id ASC")
	} else {
		tx = tx.Order("id DESC")
	}
	if err := tx.Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	for _, l := range ls {
		var statistics task.TaskStatistics
		if len(l.Statistics) > 0 {
			_ = json.Unmarshal(l.Statistics, &statistics)
		}
		ret.Rows = append(ret.Rows, task.Task{
			TaskId:     l.TaskId,
			TaskPid:    l.TaskPid,
			TaskName:   l.TaskName,
			Status:     l.Status,
			Message:    l.Message,
			Statistics: statistics,
			TaskType:   l.TaskType,
			CreatedAt:  l.CreatedAt,
			UpdatedAt:  l.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": ret})
}

func (t *Task) Retry(c *gin.Context) {
	var req task.RetryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		for _, idx := range strings.Split(req.Ids, ",") {
			var item model.Task
			if err := tx.Model(&model.Task{}).
				Where("task_id = ?", idx).
				First(&item).Error; err != nil {
				return err
			}
			switch item.TaskType {
			case service.TASK_TYPE_ARTICLE:
				otherId, err := strconv.ParseInt(item.OtherId, 10, 64)
				if err != nil {
					return err
				}
				info, err := service.GetArticleInfo(c, geek.ArticlesInfoRequest{Id: otherId})
				if err != nil {
					return err
				}
				raw, _ := json.Marshal(info.Data)
				item.Raw = raw
				item.Status = service.TASK_STATUS_PENDING
				err = tx.Model(&model.Task{}).
					Where("task_id", item.TaskId).
					Updates(&item).Error
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK"})
}
