package v2

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/middleware"
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
	identity := c.GetString(middleware.Identity)
	ret := task.TaskListResponse{
		Rows: make([]task.Task, 0, 10),
	}
	var ls []*model.Task
	tx := global.DB.Model(&model.Task{})
	if req.Xstatus > 0 {
		tx = tx.Where("status = ?", req.Xstatus)
	}
	tx = tx.Where("uid = ?", identity)
	tx = tx.Where("task_pid = ?", req.TaskPid)
	tx = tx.Where("deleted_at = ?", 0)
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
			OtherId:    l.OtherId,
			TaskName:   l.TaskName,
			Status:     l.Status,
			Statistics: statistics,
			TaskType:   l.TaskType,
			Cover:      l.Cover,
			CreatedAt:  l.CreatedAt,
			UpdatedAt:  l.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": ret})
}

func (t *Task) Info(c *gin.Context) {
	var req task.TaskInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	var l model.Task
	if err := global.DB.Model(&model.Task{}).
		Where("task_id=?", req.Id).First(&l).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	var statistics task.TaskStatistics
	if len(l.Statistics) > 0 {
		_ = json.Unmarshal(l.Statistics, &statistics)
	}

	var articleData geek.ArticleInfoResponse
	if len(l.Raw) > 0 {
		_ = json.Unmarshal(l.Raw, &articleData)
	}

	var taskMessage task.TaskMessage
	if len(l.Message) > 0 {
		_ = json.Unmarshal(l.Message, &taskMessage)
		if len(taskMessage.Object) > 0 {
			taskMessage.Object = global.Storage.GetUrl(taskMessage.Object)
		}
	}

	resp := task.TaskInfoResponse{
		Task: task.Task{
			TaskId:     l.TaskId,
			TaskPid:    l.TaskPid,
			OtherId:    l.OtherId,
			TaskName:   l.TaskName,
			Status:     l.Status,
			Cover:      l.Cover,
			Statistics: statistics,
			TaskType:   l.TaskType,
			CreatedAt:  l.CreatedAt,
			UpdatedAt:  l.UpdatedAt,
		},
		Article: articleData.Data.Info,
		Message: taskMessage,
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": resp})
}

func (t *Task) Retry(c *gin.Context) {
	var req task.RetryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Task{}).
			Where("task_id", req.Pid).
			UpdateColumn("status", service.TASK_STATUS_PENDING).Error; err != nil {
			return err
		}
		for _, idx := range strings.Split(req.Ids, ",") {
			if err := tx.Model(&model.Task{}).
				Where("task_id", idx).
				UpdateColumn("status", service.TASK_STATUS_PENDING).Error; err != nil {
				return err
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

func (t *Task) Delete(c *gin.Context) {
	var req task.DeleteRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if len(req.Pid) > 0 {
			if err := tx.Model(&model.Task{}).
				Where("task_id", req.Pid).
				Updates(map[string]any{"deleted_at": time.Now().Unix()}).Error; err != nil {
				return err
			}
			if len(req.Ids) == 0 {
				if err := tx.Model(&model.Task{}).
					Where("task_pid", req.Pid).
					Updates(map[string]any{"deleted_at": time.Now().Unix()}).Error; err != nil {
					return err
				}
			}
		}
		for _, idx := range strings.Split(req.Ids, ",") {
			if err := tx.Model(&model.Task{}).
				Where("task_id", idx).
				Updates(map[string]any{"deleted_at": time.Now().Unix()}).Error; err != nil {
				return err
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

func (t *Task) Download(c *gin.Context) {
	var req task.TaskDownloadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	var l model.Task
	if err := global.DB.Model(&model.Task{}).
		Where("task_id=?", req.Id).First(&l).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	var articleData geek.ArticleInfoResponse
	if err := json.Unmarshal(l.Raw, &articleData); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	var taskMessage task.TaskMessage
	if err := json.Unmarshal(l.Message, &taskMessage); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	baseName := service.VerifyFileName(articleData.Data.Info.Title)
	switch req.Type {
	case "markdown":
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(articleData.Data.Info.Content)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
			return
		}
		fileName := baseName + ".md"
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		c.Header("Content-Transfer-Encoding", "binary")
		c.Data(200, "application/octet-stream", []byte(markdown))
	case "audio", "video":
		fileName := baseName + ".mp4"
		if req.Type == "audio" {
			fileName = baseName + ".mp3"
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
		c.Header("Content-Transfer-Encoding", "binary")
		object := global.Storage.GetKey(taskMessage.Object, true)
		c.File(object)
	}
}
