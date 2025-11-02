package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/internal/model"
	"github.com/zkep/my-geektime/internal/service"
	"github.com/zkep/my-geektime/internal/types/collect"
	"github.com/zkep/my-geektime/internal/types/geek"
	"github.com/zkep/my-geektime/internal/types/sys_dict"
	"github.com/zkep/my-geektime/internal/types/task"
	"gorm.io/gorm"
)

type Collect struct{}

func NewCollect() *Collect {
	return &Collect{}
}

func (t *Collect) Create(c *gin.Context) {
	var req collect.CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	identity := c.GetString(global.Identity)
	ids := strings.Split(req.Ids, ",")
	switch req.CollectType {
	case collect.CollectTask:
	default:
		global.FAIL(c, "fail.msg", "暂时不支持当前类型")
		return
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			item := &model.Collect{
				Uid:         identity,
				CollectId:   id,
				CollectType: req.CollectType,
				Category:    req.Category.String(),
			}
			err := tx.Model(item).
				Where("uid = ?", identity).
				Where("collect_id = ?", id).
				Where("collect_type = ?", req.CollectType).
				FirstOrCreate(item).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	global.OK(c, nil)
}

func (t *Collect) List(c *gin.Context) {
	var req collect.CollectListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.PerPage <= 0 || (req.PerPage > 200) {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	ret := collect.CollectListResponse{
		Rows: make([]collect.Collect, 0, 10),
	}
	var ls []*model.Collect
	tx := global.DB.Model(&model.Collect{})
	if req.Category > 0 {
		tx = tx.Where("category = ?", req.Category)
	}
	tx = tx.Where("deleted_at = ?", 0)
	tx = tx.Order("updated_at DESC")
	if err := tx.Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	for _, l := range ls {
		row := collect.Collect{
			Collect: l,
		}
		if l.CollectType == collect.CollectTask {
			var x *model.Task
			if err := global.DB.Model(&model.Task{}).
				Where("task_id = ?", l.CollectId).
				Where("deleted_at = ?", 0).
				Order("id DESC").First(&x).Error; err != nil {
				// record not found, delete collect
				if errors.Is(err, gorm.ErrRecordNotFound) {
					global.DB.Where("id = ?", l.Id).Delete(&model.Collect{})
				}
				continue
			}
			var statistics task.TaskStatistics
			if len(x.Statistics) > 0 {
				_ = json.Unmarshal(x.Statistics, &statistics)
			}
			taskRow := task.Task{
				TaskId:     x.TaskId,
				TaskPid:    x.TaskPid,
				OtherId:    x.OtherId,
				OtherType:  x.OtherType,
				OtherForm:  x.OtherForm,
				OtherGroup: x.OtherGroup,
				OtherTag:   x.OtherTag,
				TaskName:   x.TaskName,
				Status:     x.Status,
				Statistics: statistics,
				TaskType:   x.TaskType,
				Cover:      x.Cover,
			}
			switch x.TaskType {
			case service.TASK_TYPE_PRODUCT:
				var product geek.ProductBase
				if len(x.Raw) > 0 {
					_ = json.Unmarshal(x.Raw, &product)
				}
				taskRow.Author = product.Author
				taskRow.Share = product.Share
				taskRow.Article = product.Article
				taskRow.Subtitle = product.Subtitle
				taskRow.IntroHTML = product.IntroHTML
				taskRow.IsVideo = product.IsVideo
				taskRow.IsAudio = product.IsAudio
				taskRow.Sale = product.Price.Sale
				taskRow.SaleType = product.Price.SaleType
				taskRow.IsAudio = product.IsAudio
				var taskMessage task.TaskMessage
				if len(x.Message) > 0 {
					_ = json.Unmarshal(x.Message, &taskMessage)
					if len(taskMessage.Object) > 0 {
						taskRow.Dir = global.Storage.GetUrl(taskMessage.Object)
						taskRow.Dir = fmt.Sprintf("%s/", taskRow.Dir)
					}
					if len(taskMessage.Doc) > 0 {
						taskRow.Doc = global.Storage.GetUrl(taskMessage.Doc)
					}
				}
				taskRow.Redirect = sys_dict.ProductURLWithType(product.Type, product.ID)
			case service.TASK_TYPE_ARTICLE:
				var articleData geek.ArticleData
				if len(x.Raw) > 0 {
					_ = json.Unmarshal(x.Raw, &articleData)
				}
				taskRow.Author = articleData.Info.Author
				taskRow.Subtitle = articleData.Info.Subtitle
				taskRow.IntroHTML = articleData.Info.Summary
				taskRow.IsVideo = articleData.Info.IsVideo
				var taskMessage task.TaskMessage
				if len(x.Message) > 0 {
					_ = json.Unmarshal(x.Message, &taskMessage)
					if len(taskMessage.Object) > 0 {
						taskRow.Object = global.Storage.GetUrl(taskMessage.Object)
					}
				}
				taskRow.Redirect = sys_dict.ProductDetailURLWithType(
					articleData.Product.Type, articleData.Info.Pid, articleData.Info.ID)
			}

			taskRow.Cover = service.URLProxyReplace(taskRow.Cover)
			taskRow.Author.Avatar = service.URLProxyReplace(taskRow.Author.Avatar)
			taskRow.Share.Cover = service.URLProxyReplace(taskRow.Share.Cover)
			if len(taskRow.IntroHTML) > 0 {
				if introHTML, err1 := service.HtmlURLProxyReplace(taskRow.IntroHTML); err1 == nil {
					taskRow.IntroHTML = introHTML
				}
			}
			row.Item, _ = json.Marshal(taskRow)
		}
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}

func (t *Collect) Delete(c *gin.Context) {
	var req collect.DeleteRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	if len(ids) > 0 {
		if err := global.DB.Where("id IN ?", ids).Delete(&model.Collect{}).Error; err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
	}
	global.OK(c, nil)
}
