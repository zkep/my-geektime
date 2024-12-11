package v2

import (
	"encoding/json"
	"fmt"
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

type Product struct{}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) List(c *gin.Context) {
	var req geek.ProductListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	req.WithLearnCount = 1
	req.Size = req.PerPage
	req.Prev = req.Page
	resp, err := service.GetLearnProduct(c, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	ret := geek.ProductListResponse{Rows: make([]geek.ProductListRow, 0)}
	ret.HasNext = resp.Data.Page.More
	for _, v := range resp.Data.Products {
		ret.Rows = append(ret.Rows, geek.ProductListRow{
			ID:            v.ID,
			Title:         v.Title,
			Subtitle:      v.Subtitle,
			Intro:         v.Intro,
			IntroHTML:     v.IntroHTML,
			Ucode:         v.Ucode,
			IsFinish:      v.IsFinish,
			IsVideo:       v.IsVideo,
			IsAudio:       v.IsAudio,
			IsDailylesson: v.IsDailylesson,
			IsUniversity:  v.IsUniversity,
			IsOpencourse:  v.IsOpencourse,
			IsQconp:       v.IsQconp,
			Share:         v.Share,
			Author:        v.Author,
			Cover:         v.Cover,
			Article:       v.Article,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": ret})
}

func (p *Product) Articles(c *gin.Context) {
	var req geek.ArticlesListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	req.Size = req.PerPage
	req.Prev = req.Page
	resp, err := service.GetArticles(c, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	ret := geek.ArticlesListResponse{Rows: make([]geek.ArticlesListRow, 0)}
	ret.Count = resp.Data.Page.Count
	for _, v := range resp.Data.List {
		item := geek.ArticlesListRow{
			ID:               v.ID,
			ArticleTitle:     v.ArticleTitle,
			ArticleSummary:   v.ArticleSummary,
			ArticleCover:     v.ArticleCover,
			VideoCover:       v.VideoCover,
			VideoSize:        v.VideoSize,
			AudioSize:        v.AudioSize,
			AudioDownloadURL: v.AudioDownloadURL,
			AuthorName:       v.AuthorName,
			AuthorIntro:      v.AuthorIntro,
		}
		if item.VideoCover != "" && item.ArticleCover == "" {
			item.ArticleCover = item.VideoCover
		}
		ret.Rows = append(ret.Rows, item)
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "data": ret})
}

func (p *Product) ArticleInfo(c *gin.Context) {
	var req geek.ArticlesInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	resp, err := service.GetArticleInfo(c, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": resp.Data.Info})
}

func (p *Product) Download(c *gin.Context) {
	var req geek.DowloadRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "error": err.Error()})
		return
	}
	ids := make([]int64, 0, 1)
	switch x := req.Ids.(type) {
	case string:
		for _, v := range strings.Split(x, ",") {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
				return
			}
			ids = append(ids, i)
		}
	case float64:
		ids = append(ids, int64(x))
	default:
		if req.Pid <= 0 {
			c.JSON(http.StatusOK, gin.H{"status": 100, "msg": "not valid ids"})
			return
		}
		resp, err := service.GetArticles(c,
			geek.ArticlesListRequest{
				Cid:   fmt.Sprintf("%d", req.Pid),
				Order: "earliest",
				Prev:  1,
				Size:  500,
			})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
			return
		}
		for _, v := range resp.Data.List {
			ids = append(ids, v.ID)
		}
	}
	raw, _ := json.Marshal(req)
	jobId := service.TaskID()
	job := &model.Task{
		TaskId:   jobId,
		TaskType: service.TASK_TYPE_PRODUCT,
		Raw:      raw,
	}
	tasks := make([]*model.Task, 0, len(ids))
	for _, id := range ids {
		info, err := service.GetArticleInfo(c, geek.ArticlesInfoRequest{Id: id})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
			return
		}
		if job.TaskName == "" {
			job.TaskName = info.Data.Product.Title
		}
		raw, _ = json.Marshal(info.Data)
		item := model.Task{
			TaskPid:  jobId,
			TaskId:   service.TaskID(),
			TaskName: info.Data.Info.Title,
			TaskType: service.TASK_TYPE_ARTICLE,
			Raw:      raw,
		}
		tasks = append(tasks, &item)
	}
	count := len(tasks)
	statistics := task.TaskStatistics{
		Count: count,
		Items: map[int]int{
			service.TASK_STATUS_PENDING:  count,
			service.TASK_STATUS_RUNNING:  0,
			service.TASK_STATUS_FINISHED: 0,
			service.TASK_STATUS_ERROR:    0,
		},
	}
	job.Statistics, _ = json.Marshal(statistics)
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(job).Error; err != nil {
			return err
		}
		for _, x := range tasks {
			if err := tx.Create(x).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": 100, "msg": err.Error()})
		return
	}
	resp := geek.DowloadResponse{JobID: jobId}
	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "OK", "data": resp})
}
