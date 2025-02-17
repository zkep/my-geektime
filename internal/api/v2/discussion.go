package v2

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
)

func (p *Product) ArticleDiscussion(c *gin.Context) {
	var req geek.DiscussionListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Prev = req.Page
	req.Size = req.PerPage
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticleCommentDiscussion(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.DiscussionListResponse{Rows: make([]geek.DiscussionData, 0)}
	ret.Count = resp.Data.Page.Total
	ret.Rows = resp.Data.List
	global.OK(c, ret)
}

func (p *Product) ArticleDiscussionList(c *gin.Context) {
	var req geek.DiscussionListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.PerPage <= 0 || (req.PerPage > 100) {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	ret := geek.DiscussionListResponse{Rows: make([]geek.DiscussionData, 0)}
	var ls []*model.ArticleCommentDiscussion
	tx := global.DB.Model(&model.ArticleCommentDiscussion{})
	if req.TargetID > 0 {
		tx = tx.Where("cid = ?", req.TargetID)
	}
	if err := tx.Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	for _, v := range ls {
		var row geek.DiscussionData
		if err := json.Unmarshal(v.Raw, &row); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}
