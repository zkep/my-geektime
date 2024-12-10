package v2

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/types/task"
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
	if req.Status > 0 {
		tx = tx.Where("status = ?", req.Status)
	}
	tx = tx.Where("task_pid = ?", req.TaskPid)
	if err := tx.Order("id DESC").
		Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	for _, l := range ls {
		var statistics task.TaskStatistics
		if len(l.Statistics) > 0 {
			if err := json.Unmarshal(l.Statistics, &statistics); err != nil {
				continue
			}
		}
		ret.Rows = append(ret.Rows, task.Task{
			TaskId:     l.TaskId,
			TaskPid:    l.TaskPid,
			TaskName:   l.TaskName,
			Status:     l.Status,
			Statistics: statistics,
			TaskType:   l.TaskType,
			CreatedAt:  l.CreatedAt,
			UpdatedAt:  l.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "data": ret})
}
