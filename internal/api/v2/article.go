package v2

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gin-gonic/gin"
	"github.com/zkep/mygeektime/internal/global"
	"github.com/zkep/mygeektime/internal/service"
	"github.com/zkep/mygeektime/internal/types/geek"
)

func (p *Product) Articles(c *gin.Context) {
	var req geek.ArticlesListRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	req.Size = req.PerPage
	req.Prev = req.Page
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticles(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	ret := geek.ArticlesListResponse{Rows: make([]geek.ArticlesListRow, 0)}
	ret.Count = resp.Data.Page.Count
	for _, v := range resp.Data.List {
		row := geek.ArticlesListRow{
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
		if row.VideoCover != "" && row.ArticleCover == "" {
			row.ArticleCover = row.VideoCover
		}
		ret.Rows = append(ret.Rows, row)
	}
	global.OK(c, ret)
}

func (p *Product) ArticleInfo(c *gin.Context) {
	var req geek.ArticlesInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	identity := c.GetString(global.Identity)
	accessToken := c.GetString(global.AccessToken)
	resp, err := service.GetArticleInfo(c, identity, accessToken, req)
	if err != nil {
		global.FAIL(c, "fail.msg", err.Error())
		return
	}
	if len(resp.Data.Info.Content) > 0 {
		if markdown, err1 := htmltomarkdown.ConvertString(resp.Data.Info.Content); err1 == nil {
			resp.Data.Info.Content = markdown
		}
	}
	global.OK(c, resp.Data.Info)
}
