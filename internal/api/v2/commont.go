package v2

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/model"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
)

func (p *Product) ArticleCommonts(c *gin.Context) {
	var req geek.ArticleCommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Prev = req.Page
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticleComment(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.ArticleCommentListResponse{Rows: make([]geek.ArticleComment, 0)}
	ret.Count = resp.Data.Page.Count
	ret.Rows = resp.Data.List
	global.OK(c, ret)
}

func (p *Product) ArticleCommontList(c *gin.Context) {
	var req geek.ArticleCommentListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	ret := geek.ArticleCommentListResponse{
		Rows: make([]geek.ArticleComment, 0, 10),
	}
	var ls []*model.ArticleComment
	tx := global.DB.Model(&model.ArticleComment{})
	if req.Aid > 0 {
		tx = tx.Where("aid = ?", req.Aid)
	}
	if err := tx.Count(&ret.Count).
		Offset((req.Page - 1) * req.PerPage).
		Limit(req.PerPage).
		Find(&ls).Error; err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	for _, v := range ls {
		var row geek.ArticleComment
		if err := json.Unmarshal(v.Raw, &row); err != nil {
			global.FAIL(c, "fail.msg", err.Error())
			return
		}
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}
